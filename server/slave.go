package server

import (
	"github.com/yudai/gotty/wetty"
)

// Slave is wetty.Slave with some additional methods.
type Slave interface {
	wetty.Slave

	Close() error
}

type Factory interface {
	Name() string
	New(params map[string][]string) (Slave, error)
}
