package server

import (
	"github.com/blevz/gotty/webtty"
)

// Upstream is webtty.Upstream with some additional methods.
type Upstream interface {
	webtty.Upstream

	Close() error
}

type Factory interface {
	Name() string
	New(params map[string][]string) (Upstream, error)
}
