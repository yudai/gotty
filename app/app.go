package app

import (
	"crypto/rand"
	"encoding/base64"
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
	"text/template"

	"github.com/braintree/manners"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/hcl"
	"github.com/kr/pty"
)

type App struct {
	command []string
	options *Options

	upgrader *websocket.Upgrader
	server   *manners.GracefulServer

	titleTemplate *template.Template
}

type Options struct {
	Address         string                 `hcl:"address"`
	Port            string                 `hcl:"port"`
	PermitWrite     bool                   `hcl:"permit_write"`
	EnableBasicAuth bool                   `hcl:"enable_basic_auth"`
	Credential      string                 `hcl:"credential"`
	EnableRandomUrl bool                   `hcl:"enable_random_url"`
	RandomUrlLength int                    `hcl:"random_url_length"`
	IndexFile       string                 `hcl:"index_file"`
	EnableTLS       bool                   `hcl:"enable_tls"`
	TLSCrtFile      string                 `hcl:"tls_crt_file"`
	TLSKeyFile      string                 `hcl:"tls_key_file"`
	TitleFormat     string                 `hcl:"title_format"`
	EnableReconnect bool                   `hcl:"enable_reconnect"`
	ReconnectTime   int                    `hcl:"reconnect_time"`
	Once            bool                   `hcl:"once"`
	Preferences     map[string]interface{} `hcl:"preferences"`
}

var DefaultOptions = Options{
	Address:         "",
	Port:            "8080",
	PermitWrite:     false,
	EnableBasicAuth: false,
	Credential:      "",
	EnableRandomUrl: false,
	RandomUrlLength: 8,
	IndexFile:       "",
	EnableTLS:       false,
	TLSCrtFile:      "~/.gotty.crt",
	TLSKeyFile:      "~/.gotty.key",
	TitleFormat:     "GoTTY - {{ .Command }} ({{ .Hostname }})",
	EnableReconnect: false,
	ReconnectTime:   10,
	Once:            false,
	Preferences:     make(map[string]interface{}),
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

	var err error
	app.server = manners.NewWithServer(
		&http.Server{Addr: endpoint, Handler: siteHandler},
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

	_, initMessage, err := conn.ReadMessage()
	if err != nil || string(initMessage) != app.options.Credential {
		log.Print("Failed to authenticate websocket connection")
		conn.Close()
		return
	}

	cmd := exec.Command(app.command[0], app.command[1:]...)
	ptyIo, err := pty.Start(cmd)
	if err != nil {
		log.Print("Failed to execute command")
		return
	}
	log.Printf("Command is running for client %s with PID %d", r.RemoteAddr, cmd.Process.Pid)

	context := &clientContext{
		app:        app,
		request:    r,
		connection: conn,
		command:    cmd,
		pty:        ptyIo,
	}

	context.goHandleClient()
}

func (app *App) handleCustomIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, ExpandHomeDir(app.options.IndexFile))
}

func (app *App) handleAuthToken(w http.ResponseWriter, r *http.Request) {
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
