package app

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/yudai/utf8reader"
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
		request:    r,
		connection: conn,
		command:    cmd,
		pty:        ptyIo,
	}

	app.goHandleConnection(context)
}

func (app *App) goHandleConnection(context *clientContext) {
	exit := make(chan bool, 2)

	go func() {
		defer func() { exit <- true }()

		app.processSend(context)
	}()

	go func() {
		defer func() { exit <- true }()

		app.processReceive(context)
	}()

	go func() {
		<-exit
		context.command.Wait()
		context.connection.Close()
		log.Printf("Connection closed: %s", context.request.RemoteAddr)
	}()
}

func (app *App) processSend(context *clientContext) {
	buf := make([]byte, 1024)
	utf8f := utf8reader.New(context.pty)

	for {
		size, err := utf8f.Read(buf)
		if err != nil {
			log.Printf("Command exited for: %s", context.request.RemoteAddr)
			return
		}

		writer, err := context.connection.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		writer.Write(buf[:size])
		writer.Close()
	}
}

func (app *App) processReceive(context *clientContext) {
	for {
		_, data, err := context.connection.ReadMessage()
		if err != nil {
			return
		}

		switch data[0] {
		case Input:
			if !app.options.PermitWrite {
				break
			}

			_, err := context.pty.Write(data[1:])
			if err != nil {
				return
			}

		case ResizeTerminal:
			var args argResizeTerminal
			err = json.Unmarshal(data[1:], &args)
			if err != nil {
				log.Print("Malformed remote command")
				return
			}

			window := struct {
				row uint16
				col uint16
				x   uint16
				y   uint16
			}{
				uint16(args.Rows),
				uint16(args.Columns),
				0,
				0,
			}
			syscall.Syscall(
				syscall.SYS_IOCTL,
				context.pty.Fd(),
				syscall.TIOCSWINSZ,
				uintptr(unsafe.Pointer(&window)),
			)

		default:
			log.Print("Unknown message type")
			return
		}
	}
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

const (
	Input          = '0'
	ResizeTerminal = '1'
)

type argResizeTerminal struct {
	Columns float64
	Rows    float64
}

type clientContext struct {
	request    *http.Request
	connection *websocket.Conn
	command    *exec.Cmd
	pty        *os.File
}
