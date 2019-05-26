package server

import (
	"fmt"
	"log"
	"net/http"
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

	// register static endpoint handlers
	mux.Handle(pathPrefix, http.StripPrefix(pathPrefix, http.FileServer(httpfs.NewFileSystem(assets.Assets, time.Now()))))

	// register ws handler
	mux.HandleFunc(pathPrefix+"ws", server.wsHandler())

	// wrap logging and compression middleware
	return handlers.LoggingHandler(os.Stderr, gziphandler.GzipHandler(mux))
}

func (server *Server) wsHandler() http.HandlerFunc {
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

		err = server.pipeWSConn(conn)

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

func (server *Server) pipeWSConn(conn *websocket.Conn) error {
	var master wetty.Master = &wsWrapper{conn}
	var slave wetty.Slave
	slave, err := server.factory.New()
	if err != nil {
		return err //ors.Wrapf(err, "failed to create backend")
	}
	defer slave.Close()

	tty, err := wetty.New(master, slave)
	if err != nil {
		return err //ors.Wrapf(err, "failed to create wetty")
	}

	return tty.Pipe()
}
