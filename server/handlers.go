package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"modernc.org/httpfs"

	"github.com/yudai/gotty/server/assets"
	"github.com/yudai/gotty/server/middleware"
	"github.com/yudai/gotty/wetty"
)

func (server *Server) setupHandlers(pathPrefix string) http.Handler {
	siteMux := http.NewServeMux()
	siteMux.HandleFunc(pathPrefix, server.handleIndex)
	siteMux.HandleFunc(pathPrefix+"auth_token.js", server.handleAuthToken)
	siteMux.HandleFunc(pathPrefix+"config.js", server.handleConfig)
	staticFileHandler := http.FileServer(httpfs.NewFileSystem(assets.Assets, time.Now()))
	siteMux.Handle(pathPrefix+"js/", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"css/", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"favicon.png", http.StripPrefix(pathPrefix, staticFileHandler))

	wsMux := http.NewServeMux()
	wsMux.Handle("/", middleware.WrapLogger(middleware.WrapGzip(http.Handler(siteMux))))
	wsMux.HandleFunc(pathPrefix+"ws", server.generateHandleWS())

	return http.Handler(wsMux)
}

func (server *Server) generateHandleWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		closeReason := "unknown reason"

		defer func() {
			log.Printf(
				"Connection closed by %s: %s, connections: %d/%d",
				closeReason, r.RemoteAddr, 0, 0,
			)
		}()

		log.Printf("New client connected: %s, connections: %d/%d", r.RemoteAddr, 0, 0)

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		conn, err := server.upgrader.Upgrade(w, r, nil)
		if err != nil {
			closeReason = err.Error()
			return
		}
		defer conn.Close()

		err = server.processWSConn(conn)

		switch err {
		case wetty.ErrSlaveClosed:
			closeReason = server.factory.Name()
		case wetty.ErrMasterClosed:
			closeReason = "client"
		default:
			closeReason = fmt.Sprintf("an error: %s", err)
		}
	}
}

func (server *Server) processWSConn(conn *websocket.Conn) error {
	typ, initLine, err := conn.ReadMessage()
	if err != nil {
		return err //ors.Wrapf(err, "failed to authenticate websocket connection")
	}
	if typ != websocket.TextMessage {
		return err //ors.New("failed to authenticate websocket connection: invalid message type")
	}

	var init InitMessage
	err = json.Unmarshal(initLine, &init)
	if err != nil {
		return err //ors.Wrapf(err, "failed to authenticate websocket connection")
	}
	if init.AuthToken != "" {
		return err //ors.New("failed to authenticate websocket connection")
	}

	queryPath := "?"
	if init.Arguments != "" {
		queryPath = init.Arguments
	}

	query, err := url.Parse(queryPath)
	if err != nil {
		return err //ors.Wrapf(err, "failed to parse arguments")
	}
	params := query.Query()
	var slave Slave
	slave, err = server.factory.New(params)
	if err != nil {
		return err //ors.Wrapf(err, "failed to create backend")
	}
	defer slave.Close()

	tty, err := wetty.New(&wsWrapper{conn}, slave)
	if err != nil {
		return err //ors.Wrapf(err, "failed to create wetty")
	}

	err = tty.Run()

	return err
}

// Dynamic: index.html
func (server *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	indexBuf := new(bytes.Buffer)
	err := server.indexTemplate.Execute(indexBuf, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.Write(indexBuf.Bytes())
}

// Dynamic: auth_token.js
func (server *Server) handleAuthToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	// @TODO hashing?
	w.Write([]byte("var gotty_auth_token = '';"))
}

// Dynamic: config.js
func (server *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte("var gotty_term = 'xterm';"))
}
