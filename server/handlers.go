package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/invctrl/hijack/wrap"
	"modernc.org/httpfs"

	"github.com/navigaid/gotty/assets"
	"github.com/navigaid/gotty/wetty"
)

func (server *Server) setupHandlers(pathPrefix string) http.Handler {
	mux := http.NewServeMux()
	// register static endpoint handlers
	mux.Handle(pathPrefix, http.StripPrefix(pathPrefix, http.FileServer(httpfs.NewFileSystem(assets.Assets, time.Now()))))
	// register ws handler
	mux.HandleFunc(pathPrefix+"ws", server.wsHandler)

	return handlers.LoggingHandler(os.Stderr, server.hijack(mux))
}

func (server *Server) hijack(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(http.CanonicalHeaderKey("Hijack")) == "true" {
			server.tcpHandler(w, r)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func (server *Server) tcpHandler(w http.ResponseWriter, r *http.Request) {
	closeReason := "unknown reason"

	defer func() {
		log.Printf(
			"Connection closed by %s: %s, connections: %d/%d",
			closeReason, r.RemoteAddr, 0, 0,
		)
	}()
	conn, err := wrap.WrapConn(w.(http.Hijacker).Hijack())
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	err = server.pipe(conn)
	switch err {
	case wetty.ErrSlaveClosed:
		closeReason = "local command"
	case wetty.ErrMasterClosed:
		closeReason = "client"
	default:
		closeReason = fmt.Sprintf("an error: %s", err)
	}
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

	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		closeReason = err.Error()
		return
	}
	defer conn.Close()

	err = server.pipe(&wsWrapper{conn})

	switch err {
	case wetty.ErrSlaveClosed:
		closeReason = "local command"
	case wetty.ErrMasterClosed:
		closeReason = "client"
	default:
		closeReason = fmt.Sprintf("an error: %s", err)
	}
}

func (server *Server) pipe(master wetty.Master) error {
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
