package server

import (
	"github.com/yudai/gotty/webtty"
)

// WithHistory represents TTY that has readable history.
// Read webtty.WithHistory documentation for more information.
type WithHistory interface {
	webtty.WithHistory
}

// Slave is webtty.Slave with some additional methods.
type Slave interface {
	webtty.Slave

	Close() error
}

type Factory interface {
	Name() string
	New(params map[string][]string) (Slave, error)
}
