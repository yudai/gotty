package webtty

import (
	"io"
)

// WithHistory represents TTY that has readable history.
// If Slave implements WithHistory interface, WebTTY calls HistoryReader() for each Slave instance on connection
// establishment will copy data from HistoryReader to Master til EOF.
// History will be sent to Master before any other data written to the slave from the moment of connection.
//
// Typical flow of new WebTTY connection:
//        [WebSocket connection establishment]
//                          V
//   [terminal initialization messages from server]
//                          V
//             [history data from server]
//                          V
//   [client<->server bidirectional data streaming]
type WithHistory interface {
	HistoryReader() io.Reader
}

// Slave represents a PTY slave, typically it's a local command.
type Slave interface {
	io.ReadWriter

	// WindowTitleVariables returns any values that can be used to fill out
	// the title of a terminal.
	WindowTitleVariables() map[string]interface{}

	// ResizeTerminal sets a new size of the terminal.
	ResizeTerminal(columns int, rows int) error
}
