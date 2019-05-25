package webtty

import (
	"io"
)

// Slave represents a PTY slave, typically it's a local command.
type Slave interface {
	io.ReadWriter

	// ResizeTerminal sets a new size of the terminal.
	ResizeTerminal(columns int, rows int) error
}
