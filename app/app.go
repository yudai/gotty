package app

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	htemplate "html/template"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	ttemplate "text/template"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

type App struct {
	options Options

	upgrader *websocket.Upgrader
}

type Options struct {
	Address     string
	Port        string
	PermitWrite bool
	Credential  string
	RandomUrl   bool
	TitleFormat string
	Command     []string
}

func New(options Options) *App {
	return &App{
		options: options,

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    []string{"gotty"},
		},
	}
}

func (app *App) Run() error {
	path := ""
	if app.options.RandomUrl {
		path += "/" + generateRandomString(8)
	}

	endpoint := app.options.Address + ":" + app.options.Port

	indexHandler := http.HandlerFunc(app.handleIndex)
	wsHandler := http.HandlerFunc(app.handleWS)
	staticHandler := http.FileServer(
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "static"},
	)

	var siteMux = http.NewServeMux()
	siteMux.Handle(path+"/", indexHandler)
	siteMux.Handle(path+"/static/", http.StripPrefix(path+"/static/", staticHandler))
	siteMux.Handle(path+"/ws", wsHandler)

	siteHandler := http.Handler(siteMux)

	if app.options.Credential != "" {
		log.Printf("Using Basic Authentication")
		siteHandler = wrapBasicAuth(siteHandler, app.options.Credential)
	}

	siteHandler = wrapLogger(siteHandler)

	log.Printf(
		"Server is starting with command: %s",
		strings.Join(app.options.Command, " "),
	)
	if app.options.Address != "" {
		log.Printf("URL: %s", "http://"+endpoint+path+"/")
	} else {
		for _, address := range listAddresses() {
			log.Printf("URL: %s", "http://"+address+":"+app.options.Port+path+"/")
		}
	}
	if err := http.ListenAndServe(endpoint, siteHandler); err != nil {
		return err
	}

	return nil
}

type TitleVars struct {
	Command  string
	Hostname string
}

type IndexVars struct {
	Title string
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	title := make([]byte, 0)
	titleBuf := bytes.NewBuffer(title)
	hostname, _ := os.Hostname()
	titleVars := TitleVars{
		Command:  strings.Join(app.options.Command, " "),
		Hostname: hostname,
	}
	titleTmpl, _ := ttemplate.New("title").Parse(app.options.TitleFormat)
	titleTmpl.Execute(titleBuf, titleVars)

	data, _ := Asset("templates/index.html")
	tmpl, _ := htemplate.New("index").Parse(string(data))

	vars := IndexVars{
		Title: titleBuf.String(),
	}

	tmpl.Execute(w, vars)
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
				addresses = append(addresses, v.IP.To16().String())
				addresses = append(addresses, v.IP.To4().String())
			}
		}
	}

	return
}
