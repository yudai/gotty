package webtty

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

// WebTTY bridges a PTY upstream and its PTY downstream.
// To support text-based streams and side channel commands such as
// terminal resizing, WebTTY uses an original protocol.
type WebTTY struct {
	// PTY downstream, which is usually a connection to browser
	downstreamWriter  Downstream
	downstreamReaders []Downstream

	// PTY upstream, usually a local tty
	upstream Upstream

	windowTitle []byte
	permitWrite bool
	columns     int
	rows        int
	reconnect   int // in seconds
	masterPrefs []byte

	bufferSize int
	writeMutex sync.Mutex
}

// New creates a new instance of WebTTY.
// downstream is a connection to the PTY downstream,
// typically it's a websocket connection to a client.
// upstream is usually a local command with a PTY.
func New(downstream Downstream, upstream Upstream, options ...Option) (*WebTTY, error) {
	wt := &WebTTY{
		downstreamWriter:  downstream,
		downstreamReaders: []Downstream{downstream},
		upstream:          upstream,

		permitWrite: false,
		columns:     0,
		rows:        0,

		bufferSize: 1024,
	}

	for _, option := range options {
		option(wt)
	}

	return wt, nil
}

func (wt *WebTTY) addDownstreamReader(downstream Downstream) {
	wt.writeMutex.Lock()
	defer wt.writeMutex.Unlock()
	wt.downstreamReaders = append(wt.downstreamReaders, downstream)
}

func (wt *WebTTY) AddDownstreamReader(ctx context.Context, downstream Downstream) error {
	err := wt.sendInitializeMessage()
	if err != nil {
		return errors.Wrapf(err, "failed to send initializing message")
	}
	wt.addDownstreamReader(downstream)
	select {
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}

// Run starts the main process of the WebTTY.
// This method blocks until the context is canceled.
// Note that the downstream and upstream are left intact even
// after the context is canceled. Closing them is caller's
// responsibility.
// If the connection to one end gets closed, returns ErrUpstreamClosed or ErrDownstreamClosed.
func (wt *WebTTY) Run(ctx context.Context) error {
	err := wt.sendInitializeMessage()
	if err != nil {
		return errors.Wrapf(err, "failed to send initializing message")
	}

	errs := make(chan error, 2)

	go func() {
		errs <- func() error {
			buffer := make([]byte, wt.bufferSize)
			for {
				n, err := wt.upstream.Read(buffer)
				if err != nil {
					return ErrUpstreamClosed
				}

				err = wt.handleUpstreamReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	go func() {
		errs <- func() error {
			buffer := make([]byte, wt.bufferSize)
			for {
				n, err := wt.downstreamWriter.Read(buffer)
				if err != nil {
					return ErrDownstreamClosed
				}

				err = wt.handleMasterReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errs:
	}

	return err
}

func (wt *WebTTY) sendInitializeMessage() error {
	err := wt.masterWrite(append([]byte{SetWindowTitle}, wt.windowTitle...))
	if err != nil {
		return errors.Wrapf(err, "failed to send window title")
	}

	if wt.reconnect > 0 {
		reconnect, _ := json.Marshal(wt.reconnect)
		err := wt.masterWrite(append([]byte{SetReconnect}, reconnect...))
		if err != nil {
			return errors.Wrapf(err, "failed to set reconnect")
		}
	}

	if wt.masterPrefs != nil {
		err := wt.masterWrite(append([]byte{SetPreferences}, wt.masterPrefs...))
		if err != nil {
			return errors.Wrapf(err, "failed to set preferences")
		}
	}

	return nil
}

func (wt *WebTTY) handleUpstreamReadEvent(data []byte) error {
	safeMessage := base64.StdEncoding.EncodeToString(data)
	err := wt.masterWrite(append([]byte{Output}, []byte(safeMessage)...))
	if err != nil {
		return errors.Wrapf(err, "failed to send message to master")
	}

	return nil
}

func (wt *WebTTY) masterWrite(data []byte) error {
	wt.writeMutex.Lock()
	defer wt.writeMutex.Unlock()
	for i, downstreamReader := range wt.downstreamReaders {
		fmt.Printf("Sending %s to %d\n", data, i)
		_, err := downstreamReader.Write(data)
		if err != nil {
			return errors.Wrapf(err, "failed to write to master")
		}
	}

	return nil
}

func (wt *WebTTY) handleMasterReadEvent(data []byte) error {
	if len(data) == 0 {
		return errors.New("unexpected zero length read from master")
	}
	fmt.Printf("Reading %s \n", string(data))

	switch data[0] {
	case Input:
		if !wt.permitWrite {
			return nil
		}

		if len(data) <= 1 {
			return nil
		}

		_, err := wt.upstream.Write(data[1:])
		if err != nil {
			return errors.Wrapf(err, "failed to write received data to upstream")
		}

	case Ping:
		err := wt.masterWrite([]byte{Pong})
		if err != nil {
			return errors.Wrapf(err, "failed to return Pong message to downstream")
		}

	case ResizeTerminal:
		if wt.columns != 0 && wt.rows != 0 {
			break
		}

		if len(data) <= 1 {
			return errors.New("received malformed remote command for terminal resize: empty payload")
		}

		var args argResizeTerminal
		err := json.Unmarshal(data[1:], &args)
		if err != nil {
			return errors.Wrapf(err, "received malformed data for terminal resize")
		}
		rows := wt.rows
		if rows == 0 {
			rows = int(args.Rows)
		}

		columns := wt.columns
		if columns == 0 {
			columns = int(args.Columns)
		}

		wt.upstream.ResizeTerminal(columns, rows)
	default:
		return errors.Errorf("unknown message type `%c`", data[0])
	}

	return nil
}

type argResizeTerminal struct {
	Columns float64
	Rows    float64
}
