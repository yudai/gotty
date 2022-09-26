package server

import (
	"github.com/gorilla/websocket"
)

type wsWrapper struct {
	*websocket.Conn
}

func (wsw *wsWrapper) Write(p []byte) (n int, err error) {
	writer, err := wsw.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	return writer.Write(p)
}

func minInt(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (wsw *wsWrapper) Read(p []byte) (n int, err error) {
	for {
		msgType, bytes, err := wsw.Conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		if msgType != websocket.TextMessage {
			continue
		}

		copy(p, bytes)
		return minInt(len(p), len(bytes)), err
	}
}
