package app

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/yudai/utf8reader"
)

type App struct {
	Address     string
	Port        string
	PermitWrite bool
	Credential  string
	Command     []string
}

func New(address string, port string, permitWrite bool, cred string, command []string) *App {
	return &App{
		Address:     address,
		Port:        port,
		PermitWrite: permitWrite,
		Credential:  cred,
		Command:     command,
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
	http.Handle("/",
		http.FileServer(
			&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "bindata"},
		),
	)
	http.HandleFunc("/ws", app.generateHandler())

	url := app.Address + ":" + app.Port
	log.Printf("Server is running at %s, command: %s", url, strings.Join(app.Command, " "))
	handler := http.Handler(http.DefaultServeMux)
	handler = loggerHandler(handler)
	if app.Credential != "" {
		handler = basicAuthHandler(handler, app.Credential)
	}
	err := http.ListenAndServe(url, handler)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) generateHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New client connected: %s", r.RemoteAddr)

		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    []string{"gotty"},
		}

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Failed to upgrade connection")
			return
		}

		cmd := exec.Command(app.Command[0], app.Command[1:]...)
		fio, err := pty.Start(cmd)
		log.Printf("Command is running for client %s with PID %d", r.RemoteAddr, cmd.Process.Pid)
		if err != nil {
			log.Print("Failed to execute command")
			return
		}

		exit := make(chan bool, 2)

		go func() {
			defer func() { exit <- true }()

			buf := make([]byte, 1024)
			utf8f := utf8reader.New(fio)

			for {
				size, err := utf8f.Read(buf)
				if err != nil {
					log.Printf("Command exited for: %s", r.RemoteAddr)
					return
				}

				writer, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}

				writer.Write(buf[:size])
				writer.Close()
			}
		}()

		go func() {
			defer func() { exit <- true }()

			for {
				_, data, err := conn.ReadMessage()
				if err != nil {
					return
				}

				switch data[0] {
				case '0':
					if !app.PermitWrite {
						break
					}

					_, err := fio.Write(data[1:])
					if err != nil {
						return
					}

				case '1':
					var remoteCmd command
					err = json.Unmarshal(data[1:], &remoteCmd)
					if err != nil {
						log.Print("Malformed remote command")
						return
					}

					switch remoteCmd.Name {
					case "resize_terminal":

						rows := remoteCmd.Arguments["rows"]
						switch rows.(type) {
						case float64:
						default:
							log.Print("Malformed remote command")
							return
						}

						cols := remoteCmd.Arguments["columns"]
						switch cols.(type) {
						case float64:
						default:
							log.Print("Malformed remote command")
							return
						}

						window := struct {
							row uint16
							col uint16
							x   uint16
							y   uint16
						}{
							uint16(rows.(float64)),
							uint16(cols.(float64)),
							0,
							0,
						}
						syscall.Syscall(
							syscall.SYS_IOCTL,
							fio.Fd(),
							syscall.TIOCSWINSZ,
							uintptr(unsafe.Pointer(&window)),
						)
					}

				default:
					log.Print("Unknown message type")
					return
				}
			}
		}()

		go func() {
			<-exit
			conn.Close()
			log.Printf("Connection closed: %s", r.RemoteAddr)
		}()
	}
}

type command struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}
