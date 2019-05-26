package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"modernc.org/httpfs"

	"github.com/yudai/gotty/server/assets"
	"github.com/yudai/gotty/wetty"
)

func (server *Server) setupHandlers(pathPrefix string) http.Handler {
	mux := http.NewServeMux()

	// register dynamic endpoint handlers
	mux.HandleFunc(pathPrefix, server.handleIndex)
	mux.HandleFunc(pathPrefix+"auth_token.js", server.handleAuthToken)
	mux.HandleFunc(pathPrefix+"config.js", server.handleConfig)

	// register static endpoint handlers
	staticFileHandler := http.FileServer(httpfs.NewFileSystem(assets.Assets, time.Now()))
	mux.Handle(pathPrefix+"js/", http.StripPrefix(pathPrefix, staticFileHandler))
	mux.Handle(pathPrefix+"css/", http.StripPrefix(pathPrefix, staticFileHandler))
	mux.Handle(pathPrefix+"favicon.png", http.StripPrefix(pathPrefix, staticFileHandler))

	// register ws handler
	mux.HandleFunc(pathPrefix+"ws", server.generateHandleWS())

	// wrap logging and compression middleware
	return handlers.LoggingHandler(os.Stderr, gziphandler.GzipHandler(mux))
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
