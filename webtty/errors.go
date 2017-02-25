package webtty

import (
	"errors"
)

var (
	ErrSlaveClosed  = errors.New("slave closed")
	ErrMasterClosed = errors.New("master closed")
)
