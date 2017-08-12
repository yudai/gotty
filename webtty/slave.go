package webtty

import (
	"io"
)

// Slave represents a PTY slave, typically it's a local command.
type Slave interface {
	io.ReadWriteCloser

	WindowTitleVariables() map[string]interface{}
	ResizeTerminal(columns int, rows int) error
}
