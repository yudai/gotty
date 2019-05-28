package utils

import (
	"io"
	"os"

	"github.com/gorilla/websocket"
)

func Expand(path string) string {
	if path[0:2] == "~/" {
		return os.Getenv("HOME") + path[1:]
	} else {
		return path
	}
}

// WsWrapper makes a io.ReadWriter from websocket.Conn, implementing the wetty.Master interface
// it is fed to wetty.New to create a WeTTY, bridging the websocket.Conn and local command
type WsWrapper struct {
	*websocket.Conn
}

func (wsw *WsWrapper) Write(p []byte) (n int, err error) {
	writer, err := wsw.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	return writer.Write(p)
}

func (wsw *WsWrapper) Read(p []byte) (n int, err error) {
	for {
		msgType, reader, err := wsw.Conn.NextReader()
		if err != nil {
			return 0, err
		}

		if msgType != websocket.TextMessage {
			continue
		}

		return reader.Read(p)
	}
}

// ReadWriter stores pointers to a Reader and a Writer.
// It implements io.ReadWriter automatically
type ReadWriter struct {
	io.Reader
	io.Writer
}
