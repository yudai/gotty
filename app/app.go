package app

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/braintree/manners"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/yudai/hcl"
	"github.com/yudai/umutex"
	"github.com/RobotsAndPencils/go-saml"
)

type InitMessage struct {
	Arguments string `json:"Arguments,omitempty"`
	AuthToken string `json:"AuthToken,omitempty"`
}

type App struct {
	command []string
	options *Options

	upgrader *websocket.Upgrader
	server   *manners.GracefulServer

	titleTemplate *template.Template

	onceMutex *umutex.UnblockingMutex
	samlUser string
}

type Options struct {
	Address             string                 `hcl:"address"`
	Port                string                 `hcl:"port"`
	PermitWrite         bool                   `hcl:"permit_write"`
	EnableBasicAuth     bool                   `hcl:"enable_basic_auth"`
	Credential          string                 `hcl:"credential"`
	EnableRandomUrl     bool                   `hcl:"enable_random_url"`
	RandomUrlLength     int                    `hcl:"random_url_length"`
	IndexFile           string                 `hcl:"index_file"`
	EnableTLS           bool                   `hcl:"enable_tls"`
	TLSCrtFile          string                 `hcl:"tls_crt_file"`
	TLSKeyFile          string                 `hcl:"tls_key_file"`
	EnableTLSClientAuth bool                   `hcl:"enable_tls_client_auth"`
	TLSCACrtFile        string                 `hcl:"tls_ca_crt_file"`
	TitleFormat         string                 `hcl:"title_format"`
	EnableReconnect     bool                   `hcl:"enable_reconnect"`
	ReconnectTime       int                    `hcl:"reconnect_time"`
	Once                bool                   `hcl:"once"`
	PermitArguments     bool                   `hcl:"permit_arguments"`
	CloseSignal         int                    `hcl:"close_signal"`
	Preferences         HtermPrefernces        `hcl:"preferences"`
	RawPreferences      map[string]interface{} `hcl:"preferences"`
	EnableSAML          bool                   `hcl:"enable_saml"`
	SAMLSSOUrl          string                 `hcl:"saml_sso_url"`
	SAMLIssuerUrl       string                 `hcl:"saml_issuer_url"`
	SAMLCrtFile         string                 `hcl:"saml_crt_file"`
	SAMLServiceHost     string                 `hcl:"saml_service_host"`
	SAMLServiceUrl      string                 `hcl:"saml_service_url"`
}

var Version = "0.0.12"

var DefaultOptions = Options{
	Address:             "",
	Port:                "8080",
	PermitWrite:         false,
	EnableBasicAuth:     false,
	Credential:          "",
	EnableRandomUrl:     false,
	RandomUrlLength:     8,
	IndexFile:           "",
	EnableTLS:           false,
	TLSCrtFile:          "~/.gotty.crt",
	TLSKeyFile:          "~/.gotty.key",
	EnableTLSClientAuth: false,
	TLSCACrtFile:        "~/.gotty.ca.crt",
	TitleFormat:         "GoTTY - {{ .Command }} ({{ .Hostname }})",
	EnableReconnect:     false,
	ReconnectTime:       10,
	Once:                false,
	CloseSignal:         1, // syscall.SIGHUP
	Preferences:         HtermPrefernces{},
	EnableSAML:         false,
	SAMLSSOUrl:          "",
	SAMLIssuerUrl:       "",
	SAMLCrtFile:         "",
	SAMLServiceHost:     "",
	SAMLServiceUrl:      "",
}

func New(command []string, options *Options) (*App, error) {
	titleTemplate, err := template.New("title").Parse(options.TitleFormat)
	if err != nil {
		return nil, errors.New("Title format string syntax error")
	}

	return &App{
		command: command,
		options: options,

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    []string{"gotty"},
		},

		titleTemplate: titleTemplate,

		onceMutex: umutex.New(),
		samlUser: "",
	}, nil
}

func ApplyConfigFile(options *Options, filePath string) error {
	filePath = ExpandHomeDir(filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	fileString := []byte{}
	log.Printf("Loading config file at: %s", filePath)
	fileString, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := hcl.Decode(options, string(fileString)); err != nil {
		return err
	}

	return nil
}

func CheckConfig(options *Options) error {
	if options.EnableTLSClientAuth && !options.EnableTLS {
		return errors.New("TLS client authentication is enabled, but TLS is not enabled")
	}
	return nil
}

func (app *App) Run() error {
	if app.options.PermitWrite {
		log.Printf("Permitting clients to write input to the PTY.")
	}

	if app.options.Once {
		log.Printf("Once option is provided, accepting only one client")
	}

	path := ""
	if app.options.EnableRandomUrl {
		path += "/" + generateRandomString(app.options.RandomUrlLength)
	}

	endpoint := net.JoinHostPort(app.options.Address, app.options.Port)

	wsHandler := http.HandlerFunc(app.handleWS)
	customIndexHandler := http.HandlerFunc(app.handleCustomIndex)
	authTokenHandler := http.HandlerFunc(app.handleAuthToken)
	samlAuthHandler := http.HandlerFunc(app.handleSamlAuth)
	staticHandler := http.FileServer(
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "static"},
	)

	var siteMux = http.NewServeMux()

	if app.options.IndexFile != "" {
		log.Printf("Using index file at " + app.options.IndexFile)
		siteMux.Handle(path+"/", customIndexHandler)
	} else {
		siteMux.Handle(path+"/", http.StripPrefix(path+"/", staticHandler))
	}
	siteMux.Handle(path+"/auth_token.js", authTokenHandler)
	siteMux.Handle(path+"/js/", http.StripPrefix(path+"/", staticHandler))
	siteMux.Handle(path+"/favicon.png", http.StripPrefix(path+"/", staticHandler))

	siteHandler := http.Handler(siteMux)

	if app.options.EnableSAML {
		siteMux.Handle(path + app.options.SAMLServiceUrl, samlAuthHandler)
	}

	if app.options.EnableBasicAuth {
		log.Printf("Using Basic Authentication")
		siteHandler = wrapBasicAuth(siteHandler, app.options.Credential)
	}

	if app.options.EnableSAML {
		log.Printf("Using SAML Authentication")
		siteHandler = wrapSAMLAuth(siteHandler, app, path)
	}

	siteHandler = wrapHeaders(siteHandler)

	wsMux := http.NewServeMux()
	wsMux.Handle("/", siteHandler)
	wsMux.Handle(path+"/ws", wsHandler)
	siteHandler = (http.Handler(wsMux))

	siteHandler = wrapLogger(siteHandler)

	scheme := "http"
	if app.options.EnableTLS {
		scheme = "https"
	}
	log.Printf(
		"Server is starting with command: %s",
		strings.Join(app.command, " "),
	)
	if app.options.Address != "" {
		log.Printf(
			"URL: %s",
			(&url.URL{Scheme: scheme, Host: endpoint, Path: path + "/"}).String(),
		)
	} else {
		for _, address := range listAddresses() {
			log.Printf(
				"URL: %s",
				(&url.URL{
					Scheme: scheme,
					Host:   net.JoinHostPort(address, app.options.Port),
					Path:   path + "/",
				}).String(),
			)
		}
	}

	server, err := app.makeServer(endpoint, &siteHandler)
	if err != nil {
		return errors.New("Failed to build server: " + err.Error())
	}
	app.server = manners.NewWithServer(
		server,
	)

	if app.options.EnableTLS {
		crtFile := ExpandHomeDir(app.options.TLSCrtFile)
		keyFile := ExpandHomeDir(app.options.TLSKeyFile)
		log.Printf("TLS crt file: " + crtFile)
		log.Printf("TLS key file: " + keyFile)

		err = app.server.ListenAndServeTLS(crtFile, keyFile)
	} else {
		err = app.server.ListenAndServe()
	}
	if err != nil {
		return err
	}

	log.Printf("Exiting...")

	return nil
}

func (app *App) makeServer(addr string, handler *http.Handler) (*http.Server, error) {
	server := &http.Server{
		Addr:    addr,
		Handler: *handler,
	}

	if app.options.EnableTLSClientAuth {
		caFile := ExpandHomeDir(app.options.TLSCACrtFile)
		log.Printf("CA file: " + caFile)
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, errors.New("Could not open CA crt file " + caFile)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("Could not parse CA crt file data in " + caFile)
		}
		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
		server.TLSConfig = tlsConfig
	}

	return server, nil
}

func (app *App) handleWS(w http.ResponseWriter, r *http.Request) {
	log.Printf("New client connected: %s", r.RemoteAddr)

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	conn, err := app.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Failed to upgrade connection: " + err.Error())
		return
	}

	_, stream, err := conn.ReadMessage()
	if err != nil {
		log.Print("Failed to authenticate websocket connection")
		conn.Close()
		return
	}
	var init InitMessage

	err = json.Unmarshal(stream, &init)
	if err != nil {
		log.Printf("Failed to parse init message %v", err)
		conn.Close()
		return
	}
	if init.AuthToken != app.options.Credential {
		log.Print("Failed to authenticate websocket connection")
		conn.Close()
		return
	}
	argv := app.command[1:]
	if app.options.PermitArguments {
		if init.Arguments == "" {
			init.Arguments = "?"
		}
		query, err := url.Parse(init.Arguments)
		if err != nil {
			log.Print("Failed to parse arguments")
			conn.Close()
			return
		}
		params := query.Query()["arg"]
		if len(params) != 0 {
			argv = append(argv, params...)
		}
	}

	app.server.StartRoutine()

	if app.options.Once {
		if app.onceMutex.TryLock() { // no unlock required, it will die soon
			log.Printf("Last client accepted, closing the listener.")
			app.server.Close()
		} else {
			log.Printf("Server is already closing.")
			conn.Close()
			return
		}
	}

	cmd := exec.Command(app.command[0], argv...)
	ptyIo, err := pty.Start(cmd)
	if err != nil {
		log.Print("Failed to execute command")
		return
	}
	log.Printf("Command is running for client %s with PID %d (args=%q)", r.RemoteAddr, cmd.Process.Pid, strings.Join(argv, " "))

	context := &clientContext{
		app:        app,
		request:    r,
		connection: conn,
		command:    cmd,
		pty:        ptyIo,
		writeMutex: &sync.Mutex{},
	}

	context.goHandleClient()
}

func (app *App) handleCustomIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, ExpandHomeDir(app.options.IndexFile))
}

func (app *App) handleAuthToken(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("var gotty_auth_token = '" + app.options.Credential + "';"))
}

func wrapSAMLAuth(handler http.Handler, app *App, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path + app.options.SAMLServiceUrl {
			handler.ServeHTTP(w, r)
			return
		}
		if app.samlUser == "" {
			http.Redirect(w, r, app.options.SAMLSSOUrl, 303)
			return
		}
		log.Printf("SAML Authentication Succeeded, access granted to: %s", app.samlUser)
		handler.ServeHTTP(w, r)
	})
}

func (app *App) handleSamlAuth(w http.ResponseWriter, r *http.Request) {
	encodedXML := r.FormValue("SAMLResponse")

	if encodedXML == "" {
		log.Printf("SAMLResponse form value missing")
		http.Error(w, "Bad Request", http.StatusUnauthorized)
		return
	}

	response, err := saml.ParseEncodedResponse(encodedXML)
	if err != nil {
		log.Printf("SAMLResponse parse: " + err.Error())
		http.Error(w, "Bad Request", http.StatusUnauthorized)
		return
	}

	sp := saml.ServiceProviderSettings{
		PublicCertPath:         app.options.TLSCrtFile,
		PrivateKeyPath:         app.options.TLSKeyFile,
		IDPSSOURL:              app.options.SAMLSSOUrl,
		IDPSSODescriptorURL:    app.options.SAMLIssuerUrl,
		IDPPublicCertPath:      app.options.SAMLCrtFile,
		SPSignRequest:          true,
		AssertionConsumerServiceURL: app.options.SAMLServiceHost + app.options.SAMLServiceUrl,
	}
	sp.Init()

	err = response.Validate(&sp)
	if err != nil {
		log.Printf("SAMLResponse validation: " + err.Error())
		http.Error(w, "Bad Request", http.StatusUnauthorized)
		return
	}

	samlID := response.GetAttribute("uid")
	if samlID == "" {
		log.Printf("SAML attribute identifier uid missing")
		http.Error(w, "Bad Request", http.StatusUnauthorized)
		return
	}

	app.samlUser = samlID
	http.Redirect(w, r, "/", 303)
}

func (app *App) Exit() (firstCall bool) {
	if app.server != nil {
		firstCall = app.server.Close()
		if firstCall {
			log.Printf("Received Exit command, waiting for all clients to close sessions...")
		}
		return firstCall
	}
	return true
}

func wrapLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWrapper{w, 200}
		handler.ServeHTTP(rw, r)
		log.Printf("%s %d %s %s", r.RemoteAddr, rw.status, r.Method, r.URL.Path)
	})
}

func wrapHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "GoTTY/"+Version)
		handler.ServeHTTP(w, r)
	})
}

func wrapBasicAuth(handler http.Handler, credential string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(token) != 2 || strings.ToLower(token[0]) != "basic" {
			w.Header().Set("WWW-Authenticate", `Basic realm="GoTTY"`)
			http.Error(w, "Bad Request", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(token[1])
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if credential != string(payload) {
			w.Header().Set("WWW-Authenticate", `Basic realm="GoTTY"`)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		log.Printf("Basic Authentication Succeeded: %s", r.RemoteAddr)
		handler.ServeHTTP(w, r)
	})
}

func generateRandomString(length int) string {
	const base = 36
	size := big.NewInt(base)
	n := make([]byte, length)
	for i, _ := range n {
		c, _ := rand.Int(rand.Reader, size)
		n[i] = strconv.FormatInt(c.Int64(), base)[0]
	}
	return string(n)
}

func listAddresses() (addresses []string) {
	ifaces, _ := net.Interfaces()

	addresses = make([]string, 0, len(ifaces))

	for _, iface := range ifaces {
		ifAddrs, _ := iface.Addrs()
		for _, ifAddr := range ifAddrs {
			switch v := ifAddr.(type) {
			case *net.IPNet:
				addresses = append(addresses, v.IP.String())
			case *net.IPAddr:
				addresses = append(addresses, v.IP.String())
			}
		}
	}

	return
}

func ExpandHomeDir(path string) string {
	if path[0:2] == "~/" {
		return os.Getenv("HOME") + path[1:]
	} else {
		return path
	}
}
