package webtty

import (
	"io"
)

// Downstream represents a PTY sending input and receiving output from the Upstream, usually it's a websocket connection.
type Downstream io.ReadWriter
