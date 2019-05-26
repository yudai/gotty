package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"modernc.org/httpfs"

	"github.com/yudai/gotty/assets"
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
			closeReason = "local command"
		case wetty.ErrMasterClosed:
			closeReason = "client"
		default:
			closeReason = fmt.Sprintf("an error: %s", err)
		}
	}
}

func (server *Server) parseInitMessage(conn *websocket.Conn) (map[string][]string, error) {
	typ, initLine, err := conn.ReadMessage()
	if err != nil {
		return nil, err //ors.Wrapf(err, "failed to authenticate websocket connection")
	}
	if typ != websocket.TextMessage {
		return nil, errors.New("failed to authenticate websocket connection: invalid message type")
	}

	type InitMessage struct {
		Arguments string `json:"Arguments,omitempty"`
	}

	var init InitMessage
	err = json.Unmarshal(initLine, &init)
	if err != nil {
		return nil, err //ors.Wrapf(err, "failed to authenticate websocket connection")
	}

	queryPath := "?"
	if init.Arguments != "" {
		queryPath = init.Arguments
	}

	query, err := url.Parse(queryPath)
	if err != nil {
		return nil, err //ors.Wrapf(err, "failed to parse arguments")
	}
	return query.Query(), nil
}

func (server *Server) processWSConn(conn *websocket.Conn) error {
	params, err := server.parseInitMessage(conn)
	if err != nil {
		return err
	}

	var master wetty.Master = &wsWrapper{conn}
	var slave wetty.Slave
	slave, err = server.factory.New(params)
	if err != nil {
		return err //ors.Wrapf(err, "failed to create backend")
	}
	defer slave.Close()

	tty, err := wetty.New(master, slave)
	if err != nil {
		return err //ors.Wrapf(err, "failed to create wetty")
	}

	return tty.Run()
}

// Dynamic: index.html
func (server *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	indexData, ok := assets.Assets["/index.html"]
	if !ok {
		log.Fatalln("index not found") // must be in assets.Assets
	}

	indexTemplate, err := template.New("index").Parse(indexData)
	if err != nil {
		log.Fatalln("index template parse failed") // must be valid
	}

	indexBuf := new(bytes.Buffer)
	if err := indexTemplate.Execute(indexBuf, nil); err != nil {
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
