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
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/yudai/gotty/utils"

	"github.com/braintree/manners"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/yudai/umutex"
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
	timer     *time.Timer

	// clientContext writes concurrently
	// Use atomic operations.
	connections *int64
}

type Options struct {
	Address             string                 `hcl:"address" flagName:"address" flagSName:"a" flagDescribe:"IP address to listen" default:""`
	Port                string                 `hcl:"port" flagName:"port" flagSName:"p" flagDescribe:"Port number to liten" default:"8080"`
	PermitWrite         bool                   `hcl:"permit_write" flagName:"permit-write" flagSName:"w" flagDescribe:"Permit clients to write to the TTY (BE CAREFUL)" default:"false"`
	EnableBasicAuth     bool                   `hcl:"enable_basic_auth" default:"false"`
	Credential          string                 `hcl:"credential" flagName:"credential" flagSName:"c" flagDescribe:"Credential for Basic Authentication (ex: user:pass, default disabled)" default:""`
	EnableRandomUrl     bool                   `hcl:"enable_random_url flagName:"random-url" flagSName:"r" flagDescribe:"Add a random string to the URL"" default:"false"`
	RandomUrlLength     int                    `hcl:"random_url_length" flagName:"random-url-length" flagDescribe:"Random URL length" default:"8"`
	IndexFile           string                 `hcl:"index_file" flagName:"index" flagDescribe:"Custom index.html file" default:""`
	EnableTLS           bool                   `hcl:"enable_tls" flagName:"tls" flagSName:"t" flagDescribe:"Enable TLS/SSL" default:"false"`
	TLSCrtFile          string                 `hcl:"tls_crt_file" flagName:"tls-crt" flagDescribe:"TLS/SSL certificate file path" default:"~/.gotty.crt"`
	TLSKeyFile          string                 `hcl:"tls_key_file" flagName:"tls-key" flagDescribe:"TLS/SSL key file path" default:"~/.gotty.key"`
	EnableTLSClientAuth bool                   `hcl:"enable_tls_client_auth" default:"false"`
	TLSCACrtFile        string                 `hcl:"tls_ca_crt_file" flagName:"tls-ca-crt" flagDescribe:"TLS/SSL CA certificate file for client certifications" default:"~/.gotty.ca.crt"`
	TitleFormat         string                 `hcl:"title_format" flagName:"title-format" flagDescribe:"Title format of browser window" default:"GoTTY - {{ .Command }} ({{ .Hostname }})"`
	EnableReconnect     bool                   `hcl:"enable_reconnect" flagName:"reconnect" flagDescribe:"Enable reconnection" default:"false"`
	ReconnectTime       int                    `hcl:"reconnect_time" flagName:"reconnect-time" flagDescribe:"Time to reconnect" default:"10"`
	MaxConnection       int                    `hcl:"max_connection" flagName:"max-connection" flagDescribe:"Maximum connection to gotty" default:"0"`
	Once                bool                   `hcl:"once" flagName:"once" flagDescribe:"Accept only one client and exit on disconnection" default:"false"`
	Timeout             int                    `hcl:"timeout" flagName:"timeout" flagDescribe:"Timeout seconds for waiting a client(0 to disable)" default:"0"`
	PermitArguments     bool                   `hcl:"permit_arguments" flagName:"permit-arguments" flagDescribe:"Permit clients to send command line arguments in URL (e.g. http://example.com:8080/?arg=AAA&arg=BBB)" default:"true"`
	CloseSignal         int                    `hcl:"close_signal" flagName:"close-signal" flagDescribe:"Signal sent to the command process when gotty close it (default: SIGHUP)" default:"1"`
	Preferences         HtermPrefernces        `hcl:"preferences"`
	RawPreferences      map[string]interface{} `hcl:"preferences"`
	Width               int                    `hcl:"width" flagName:"width" flagDescribe:"Static width of the screen, 0(default) means dynamically resize" default:"0"`
	Height              int                    `hcl:"height" flagName:"height" flagDescribe:"Static height of the screen, 0(default) means dynamically resize" default:"0"`
}

var Version = "0.0.13"

func New(command []string, options *Options) (*App, error) {
	titleTemplate, err := template.New("title").Parse(options.TitleFormat)
	if err != nil {
		return nil, errors.New("Title format string syntax error")
	}

	connections := int64(0)

	return &App{
		command: command,
		options: options,

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    []string{"gotty"},
		},

		titleTemplate: titleTemplate,

		onceMutex:   umutex.New(),
		connections: &connections,
	}, nil
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

	if app.options.EnableBasicAuth {
		log.Printf("Using Basic Authentication")
		siteHandler = wrapBasicAuth(siteHandler, app.options.Credential)
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

	if app.options.Timeout > 0 {
		app.timer = time.NewTimer(time.Duration(app.options.Timeout) * time.Second)
		go func() {
			<-app.timer.C
			app.Exit()
		}()
	}

	if app.options.EnableTLS {
		crtFile := utils.ExpandHomeDir(app.options.TLSCrtFile)
		keyFile := utils.ExpandHomeDir(app.options.TLSKeyFile)
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
		caFile := utils.ExpandHomeDir(app.options.TLSCACrtFile)
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

func (app *App) stopTimer() {
	if app.options.Timeout > 0 {
		app.timer.Stop()
	}
}

func (app *App) restartTimer() {
	if app.options.Timeout > 0 {
		app.timer.Reset(time.Duration(app.options.Timeout) * time.Second)
	}
}

func (app *App) handleWS(w http.ResponseWriter, r *http.Request) {
	app.stopTimer()

	connections := atomic.AddInt64(app.connections, 1)
	if int64(app.options.MaxConnection) != 0 {
		if connections >= int64(app.options.MaxConnection) {
			log.Printf("Reached max connection: %d", app.options.MaxConnection)
			return
		}
	}
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

	if app.options.MaxConnection != 0 {
		log.Printf("Command is running for client %s with PID %d (args=%q), connections: %d/%d",
			r.RemoteAddr, cmd.Process.Pid, strings.Join(argv, " "), connections, app.options.MaxConnection)
	} else {
		log.Printf("Command is running for client %s with PID %d (args=%q), connections: %d",
			r.RemoteAddr, cmd.Process.Pid, strings.Join(argv, " "), connections)
	}

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
	http.ServeFile(w, r, utils.ExpandHomeDir(app.options.IndexFile))
}

func (app *App) handleAuthToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte("var gotty_auth_token = '" + app.options.Credential + "';"))
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
