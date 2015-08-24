package app

import (
	"crypto/rand"
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
	"text/template"

	"github.com/braintree/manners"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

type App struct {
	options Options

	upgrader *websocket.Upgrader
	server   *manners.GracefulServer

	preferences   map[string]interface{}
	titleTemplate *template.Template
}

type Options struct {
	Address       string
	Port          string
	PermitWrite   bool
	Credential    string
	RandomUrl     bool
	ProfileFile   string
	EnableTLS     bool
	TLSCert       string
	TLSKey        string
	TitleFormat   string
	AutoReconnect int
	Command       []string
}

const DefaultProfileFilePath = "~/.gotty"
const DefaultTLSKeyPath = "~/.gotty.key"
const DefaultTLSCertPath = "~/.gotty.crt"

func New(options Options) (*App, error) {
	titleTemplate, err := template.New("title").Parse(options.TitleFormat)
	if err != nil {
		return nil, errors.New("Title format string syntax error")
	}

	prefMap, err := loadProfileFile(options)
	if err != nil {
		return nil, err
	}

	return &App{
		options: options,

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    []string{"gotty"},
		},

		preferences:   prefMap,
		titleTemplate: titleTemplate,
	}, nil
}

func loadProfileFile(options Options) (map[string]interface{}, error) {
	prefString := []byte{}
	prefPath := options.ProfileFile
	if options.ProfileFile == DefaultProfileFilePath {
		prefPath = os.Getenv("HOME") + "/.gotty"
	}
	if _, err := os.Stat(prefPath); os.IsNotExist(err) {
		if options.ProfileFile != DefaultProfileFilePath {
			return nil, err
		}
	} else {
		log.Printf("Loading profile path: %s", prefPath)
		prefString, _ = ioutil.ReadFile(prefPath)
	}
	if len(prefString) == 0 {
		prefString = []byte(("{}"))
	}
	var prefMap map[string]interface{}
	err := json.Unmarshal(prefString, &prefMap)
	if err != nil {
		return nil, err
	}
	return prefMap, nil
}

func (app *App) Run() error {
	if app.options.PermitWrite {
		log.Printf("Permitting clients to write input to the PTY.")
	}

	path := ""
	if app.options.RandomUrl {
		path += "/" + generateRandomString(8)
	}

	endpoint := net.JoinHostPort(app.options.Address, app.options.Port)

	wsHandler := http.HandlerFunc(app.handleWS)
	staticHandler := http.FileServer(
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "static"},
	)

	var siteMux = http.NewServeMux()
	siteMux.Handle(path+"/", http.StripPrefix(path+"/", staticHandler))
	siteMux.Handle(path+"/ws", wsHandler)

	siteHandler := http.Handler(siteMux)

	if app.options.Credential != "" {
		log.Printf("Using Basic Authentication")
		siteHandler = wrapBasicAuth(siteHandler, app.options.Credential)
	}

	siteHandler = wrapLogger(siteHandler)

	scheme := "http"
	if app.options.EnableTLS {
		scheme = "https"
	}
	log.Printf(
		"Server is starting with command: %s",
		strings.Join(app.options.Command, " "),
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
		cert, key := app.loadTLSFiles()
		err = app.server.ListenAndServeTLS(cert, key)
	} else {
		err = app.server.ListenAndServe()
	}
	if err != nil {
		return err
	}

	log.Printf("Exiting...")

	return nil
}

func (app *App) loadTLSFiles() (cert string, key string) {
	cert = app.options.TLSCert
	if app.options.TLSCert == DefaultTLSCertPath {
		cert = os.Getenv("HOME") + "/.gotty.crt"
	}

	key = app.options.TLSKey
	if app.options.TLSKey == DefaultTLSKeyPath {
		key = os.Getenv("HOME") + "/.gotty.key"
	}

	return
}

func (app *App) handleWS(w http.ResponseWriter, r *http.Request) {
	log.Printf("New client connected: %s", r.RemoteAddr)

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	conn, err := app.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Failed to upgrade connection")
		return
	}

	cmd := exec.Command(app.options.Command[0], app.options.Command[1:]...)
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

func (app *App) Exit() (firstCall bool) {
	if app.server != nil {
		log.Printf("Received Exit command, waiting for all clients to close sessions...")
		return app.server.Close()
	}
	return true
}

func wrapLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
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
