package app

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"math/big"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

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

func loggerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

func basicAuthHandler(h http.Handler, cred string) http.Handler {
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
		if cred != string(payload) {
			w.Header().Set("WWW-Authenticate", `Basic realm="GoTTY"`)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (app *App) Run() error {
	path := "/"
	if app.options.RandomUrl {
		randomPath := generateRandomString(8)
		path = "/" + randomPath + "/"
	}

	fs := http.StripPrefix(path, http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "bindata"}))
	http.Handle(path, fs)
	http.HandleFunc(path+"ws", app.handler)

	endpoint := app.options.Address + ":" + app.options.Port
	log.Printf("Server is running at %s, command: %s", endpoint+path, strings.Join(app.options.Command, " "))
	handler := http.Handler(http.DefaultServeMux)
	handler = loggerHandler(handler)
	if app.options.Credential != "" {
		handler = basicAuthHandler(handler, app.options.Credential)
	}
	err := http.ListenAndServe(endpoint, handler)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) handler(w http.ResponseWriter, r *http.Request) {
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
