// Package webtty provides a protocl and an implementation to
// controll terminals thorough networks.
package wetty

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

type (
	// Master represents a PTY master, usually it's a websocket connection.
	Master interface {
		io.ReadWriter
	}

	// Slave represents a PTY slave, typically it's a local command.
	Slave interface {
		io.ReadWriteCloser
		ResizeTerminal(columns int, rows int) error
	}

	// WeTTY bridges a PTY slave and its PTY master.
	// To support text-based streams and side channel commands such as
	// terminal resizing, WeTTY uses an original protocol.
	WeTTY struct {
		mu     sync.Mutex // guards writes to master
		master Master     // PTY Master, which probably a connection to browser
		slave  Slave      // PTY Slave

		bufferSize int
	}
)

///webtty.go

// New creates a new instance of WeTTY.
// masterConn is a connection to the PTY master,
// typically it's a websocket connection to a client.
// slave is a PTY slave such as a local command with a PTY.
func New(master Master, slave Slave) (*WeTTY, error) {
	wt := &WeTTY{
		master: master,
		slave:  slave,

		bufferSize: 1024,
	}

	return wt, nil
}

// Run starts the main process of the WeTTY.
// This method blocks until the context is canceled.
// Note that the master and slave are left intact even
// after the context is canceled. Closing them is caller's
// responsibility.
// If the connection to one end gets closed, returns ErrSlaveClosed or ErrMasterClosed.
func (wt *WeTTY) Pipe() error {
	errs := make(chan error, 2)

	// slave => buffer => master
	go func() {
		errs <- func() error {
			buffer := make([]byte, wt.bufferSize)
			for {
				n, err := wt.slave.Read(buffer)
				if err != nil {
					return ErrSlaveClosed
				}

				err = wt.handleSlaveReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	// slave <= buffer <= master
	go func() {
		errs <- func() error {
			buffer := make([]byte, wt.bufferSize)
			for {
				n, err := wt.master.Read(buffer)
				if err != nil {
					return ErrMasterClosed
				}

				err = wt.handleMasterReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	return <-errs
}

func (wt *WeTTY) handleSlaveReadEvent(data []byte) error {
	safeMessage := base64.StdEncoding.EncodeToString(data)
	return wt.masterWrite(append([]byte{Output}, []byte(safeMessage)...))
}

func (wt *WeTTY) masterWrite(data []byte) error {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	_, err := wt.master.Write(data)
	return err
}

func (wt *WeTTY) handleMasterReadEvent(data []byte) error {
	if len(data) == 0 {
		return errors.New("unexpected zero length read from master")
	}

	switch data[0] {
	case Input:
		if len(data) <= 1 {
			return nil
		}

		_, err := wt.slave.Write(data[1:])
		if err != nil {
			return err //ors.Wrapf(err, "failed to write received data to slave")
		}

	case Ping:
		err := wt.masterWrite([]byte{Pong})
		if err != nil {
			return err //ors.Wrapf(err, "failed to return Pong message to master")
		}

	case ResizeTerminal:
		type termSize struct {
			Columns float64
			Rows    float64
		}
		var args termSize
		if len(data) <= 1 {
			return errors.New("received malformed remote command for terminal resize: empty payload")
		}
		if err := json.Unmarshal(data[1:], &args); err != nil {
			return err //ors.Wrapf(err, "received malformed data for terminal resize")
		}
		wt.slave.ResizeTerminal(int(args.Rows), int(args.Columns))
	default:
		return errors.New(fmt.Sprintf("unknown message type `%c`", data[0]))
	}

	return nil
}

///protocol.go

// Protocols defines the name of this protocol,
// which is supposed to be used to the subprotocol of Websockt streams.
var Protocols = []string{"webtty"}

// those message types are used when reading from master
const (
	// Unknown message type, maybe sent by a bug
	UnknownInput = '0'
	// User input typically from a keyboard
	Input = '1'
	// Ping to the server from master(browser)
	Ping = '2'
	// Notify the server that the browser size has been changed, the server then tries to resize the slave tty
	ResizeTerminal = '3'
)

// those message types are used when writing to master
const (
	// Unknown message type, maybe set by a bug
	UnknownOutput = '0'
	// Normal output from the slave tty
	Output = '1'
	// Pong to the master(browser)
	Pong = '2'
)

///errors.go

var (
	// ErrSlaveClosed indicates the function has exited by the slave
	ErrSlaveClosed = errors.New("slave closed")

	// ErrSlaveClosed is returned when the slave connection is closed.
	ErrMasterClosed = errors.New("master closed")
)
