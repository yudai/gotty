package backends

import (
	"io"
	"net/url"
)

type ClientContextManager interface {
	New(params url.Values) (ClientContext, error)
}

type ClientContext interface {
	WindowTitle() (string, error)
	Start(exitCh chan bool)
	InputWriter() io.Writer
	OutputReader() io.Reader
	ResizeTerminal(width, height uint16) error
	TearDown() error
}
