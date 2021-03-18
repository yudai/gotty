package webtty

import (
	"errors"
)

var (
	// ErrUpstreamClosed indicates the function has exited by the upstream
	ErrUpstreamClosed = errors.New("upstream closed")

	// ErrDownstreamClosed is returned when the downstream connection is closed.
	ErrDownstreamClosed = errors.New("downstream closed")
)
