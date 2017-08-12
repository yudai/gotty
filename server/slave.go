package server

import (
	"github.com/yudai/gotty/webtty"
)

// Slave is webtty.Slave with some additional methods.
type Slave interface {
	webtty.Slave

	GetTerminalSize() (width int, height int, err error)
}

type Factory interface {
	Name() string
	New(params map[string][]string) (Slave, error)
}
