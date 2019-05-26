package server

import (
	"github.com/yudai/gotty/wetty"
)

// Slave is wetty.Slave with some additional methods.
type Slave interface {
	wetty.Slave

	Close() error
}
