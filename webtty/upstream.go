package webtty

import (
	"io"
)

// Upstream represents a PTY doing the executing, typically it's a local command.
type Upstream interface {
	io.ReadWriter

	// WindowTitleVariables returns any values that can be used to fill out
	// the title of a terminal.
	WindowTitleVariables() map[string]interface{}

	// ResizeTerminal sets a new size of the terminal.
	ResizeTerminal(columns int, rows int) error
}
