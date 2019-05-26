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

///errors.go

var (
	// ErrSlaveClosed indicates the function has exited by the slave
	ErrSlaveClosed = errors.New("slave closed")

	// ErrSlaveClosed is returned when the slave connection is closed.
	ErrMasterClosed = errors.New("master closed")
)

///master.go

// Master represents a PTY master, usually it's a websocket connection.
type Master io.ReadWriter

///message_types.go

// Protocols defines the name of this protocol,
// which is supposed to be used to the subprotocol of Websockt streams.
var Protocols = []string{"webtty"}

const (
	// Unknown message type, maybe sent by a bug
	UnknownInput = '0'
	// User input typically from a keyboard
	Input = '1'
	// Ping to the server
	Ping = '2'
	// Notify that the browser size has been changed
	ResizeTerminal = '3'
)

const (
	// Unknown message type, maybe set by a bug
	UnknownOutput = '0'
	// Normal output to the terminal
	Output = '1'
	// Pong to the browser
	Pong = '2'
	// Set window title of the terminal
	SetWindowTitle = '3'
	// Set terminal preference
	SetPreferences = '4'
	// Make terminal to reconnect
	SetReconnect = '5'
)

///slave.go

// Slave represents a PTY slave, typically it's a local command.
type Slave interface {
	io.ReadWriter

	// ResizeTerminal sets a new size of the terminal.
	ResizeTerminal(columns int, rows int) error
}

///webtty.go

// WeTTY bridges a PTY slave and its PTY master.
// To support text-based streams and side channel commands such as
// terminal resizing, WeTTY uses an original protocol.
type WeTTY struct {
	// PTY Master, which probably a connection to browser
	masterConn Master
	// PTY Slave
	slave Slave

	windowTitle []byte
	permitWrite bool
	columns     int
	rows        int
	reconnect   int // in seconds
	masterPrefs []byte

	bufferSize int
	writeMutex sync.Mutex
}

// New creates a new instance of WeTTY.
// masterConn is a connection to the PTY master,
// typically it's a websocket connection to a client.
// slave is a PTY slave such as a local command with a PTY.
func New(masterConn Master, slave Slave) (*WeTTY, error) {
	wt := &WeTTY{
		masterConn: masterConn,
		slave:      slave,

		permitWrite: true,
		columns:     0,
		rows:        0,

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
func (wt *WeTTY) Run() error {
	err := wt.sendInitializeMessage()
	if err != nil {
		return err //errors.Wrapf(err, "failed to send initializing message")
	}

	errs := make(chan error, 2)

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

	go func() {
		errs <- func() error {
			buffer := make([]byte, wt.bufferSize)
			for {
				n, err := wt.masterConn.Read(buffer)
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

	select {
	case err = <-errs:
	}

	return err
}

func (wt *WeTTY) sendInitializeMessage() error {
	err := wt.masterWrite(append([]byte{SetWindowTitle}, wt.windowTitle...))
	if err != nil {
		return err // ors.Wrapf(err, "failed to send window title")
	}

	if wt.reconnect > 0 {
		reconnect, _ := json.Marshal(wt.reconnect)
		err := wt.masterWrite(append([]byte{SetReconnect}, reconnect...))
		if err != nil {
			return err //ors.Wrapf(err, "failed to set reconnect")
		}
	}

	if wt.masterPrefs != nil {
		err := wt.masterWrite(append([]byte{SetPreferences}, wt.masterPrefs...))
		if err != nil {
			return err //ors.Wrapf(err, "failed to set preferences")
		}
	}

	return nil
}

func (wt *WeTTY) handleSlaveReadEvent(data []byte) error {
	safeMessage := base64.StdEncoding.EncodeToString(data)
	err := wt.masterWrite(append([]byte{Output}, []byte(safeMessage)...))
	if err != nil {
		return err // ors.Wrapf(err, "failed to send message to master")
	}

	return nil
}

func (wt *WeTTY) masterWrite(data []byte) error {
	wt.writeMutex.Lock()
	defer wt.writeMutex.Unlock()

	_, err := wt.masterConn.Write(data)
	if err != nil {
		return err //ors.Wrapf(err, "failed to write to master")
	}

	return nil
}

func (wt *WeTTY) handleMasterReadEvent(data []byte) error {
	if len(data) == 0 {
		return errors.New("unexpected zero length read from master")
	}

	switch data[0] {
	case Input:
		if !wt.permitWrite {
			return nil
		}

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
		if wt.columns != 0 && wt.rows != 0 {
			break
		}

		if len(data) <= 1 {
			return errors.New("received malformed remote command for terminal resize: empty payload")
		}

		var args argResizeTerminal
		err := json.Unmarshal(data[1:], &args)
		if err != nil {
			return err //ors.Wrapf(err, "received malformed data for terminal resize")
		}
		rows := wt.rows
		if rows == 0 {
			rows = int(args.Rows)
		}

		columns := wt.columns
		if columns == 0 {
			columns = int(args.Columns)
		}

		wt.slave.ResizeTerminal(columns, rows)
	default:
		return errors.New(fmt.Sprintf("unknown message type `%c`", data[0]))
	}

	return nil
}

type argResizeTerminal struct {
	Columns float64
	Rows    float64
}
