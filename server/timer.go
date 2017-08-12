package server

import (
	"time"
)

func (server *Server) stopTimer() {
	if server.options.Timeout > 0 {
		server.timer.Stop()
	}
}

func (server *Server) resetTimer() {
	if server.options.Timeout > 0 {
		server.timer.Reset(time.Duration(server.options.Timeout) * time.Second)
	}
}
