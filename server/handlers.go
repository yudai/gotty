package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"modernc.org/httpfs"

	"github.com/navigaid/gotty/assets"
	"github.com/navigaid/gotty/utils"
	"github.com/navigaid/gotty/wetty"
)

func (server *Server) setupHandlers(pathPrefix string) http.Handler {
	mux := http.NewServeMux()
	staticFileServer := http.FileServer(httpfs.NewFileSystem(assets.Assets, time.Now()))
	mux.Handle(pathPrefix, http.StripPrefix(pathPrefix, staticFileServer))
	mux.HandleFunc(pathPrefix+"ws", server.wsHandler)
	return handlers.LoggingHandler(os.Stderr, mux)
}

func (server *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
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

	conn, err := wetty.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		closeReason = err.Error()
		return
	}
	defer conn.Close()

	var master wetty.Master = &utils.WsWrapper{conn}
	var slave wetty.Slave
	slave, err = server.factory.New()
	if err != nil {
		closeReason = "slave creation failed"
		return
	}
	defer slave.Close()
	err = wetty.NewMSPair(master, slave).Pipe()
	switch err {
	case wetty.ErrSlaveClosed:
		closeReason = "local command"
	case wetty.ErrMasterClosed:
		closeReason = "client"
	default:
		closeReason = fmt.Sprintf("an error: %s", err)
	}
}
