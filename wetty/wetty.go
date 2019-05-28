// Package webtty provides a protocl and an implementation to
// controll terminals thorough networks.
package wetty

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

type (
	// Client represents the actual PTY, be it xterm.js or whatever
	Client interface {
		io.ReadWriter
	}

	// Master represents a PTY master, usually it's a websocket connection.
	Master interface {
		io.ReadWriter
	}

	// Slave represents a PTY slave, typically it's a local command.
	Slave interface {
		io.ReadWriteCloser
		ResizeTerminal(*pty.Winsize) error
	}

	CMPair struct {
		mu         sync.Mutex // guard writes to client
		mu2        sync.Mutex
		client     Client
		master     Master
		bufferSize int
	}

	// WeTTY bridges a PTY slave and its PTY master.
	// To support text-based streams and side channel commands such as
	// terminal resizing, WeTTY uses an original protocol.
	MSPair struct {
		mu         sync.Mutex // guards writes to master
		mu2        sync.Mutex // guards writes to slave
		master     Master     // PTY Master, which probably a connection to browser
		slave      Slave      // PTY Slave
		bufferSize int
	}
)

func NewCMPair(client Client, master Master) *CMPair {
	return &CMPair{
		master:     master,
		client:     client,
		bufferSize: 1024,
	}
}

func (cm *CMPair) Pipe() error {
	errs := make(chan error, 2)

	// master => buffer => client
	go func() {
		errs <- func() error {
			buffer := make([]byte, cm.bufferSize)
			for {
				n, err := cm.master.Read(buffer)
				if err != nil {
					return ErrMasterClosed
				}

				err = cm.handleMasterReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	// master <= buffer <= client
	go func() {
		errs <- func() error {
			buffer := make([]byte, cm.bufferSize)
			for {
				n, err := cm.client.Read(buffer)
				if err != nil {
					return ErrClientClosed
				}

				err = cm.handleClientReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	// master <= ping <= client
	go func() {
		errs <- func() error {
			for range time.Tick(time.Second) {
				if err := cm.masterWrite([]byte{Ping}); err != nil {
					return ErrMasterClosed
				}
			}
			return nil
		}()
	}()

	return <-errs
}

func (cm *CMPair) handleMasterReadEvent(data []byte) error {
	if len(data) == 0 {
		return errors.New("unexpected zero length read from master")
	}
	switch data[0] {
	case Pong:
	case Output:
		if len(data) <= 1 {
			return nil
		}
		if err := cm.clientWrite(data[1:]); err != nil {
			return err //ors.Wrapf(err, "failed to write received data to slave")
		}
	default:
		return errors.New(fmt.Sprintf("unknown message type `%c`", data[0]))
	}
	return nil
}

func (cm *CMPair) clientWrite(data []byte) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	_, err := cm.client.Write(data)
	return err
}

func (cm *CMPair) masterWrite(data []byte) error {
	cm.mu2.Lock()
	defer cm.mu2.Unlock()
	_, err := cm.master.Write(data)
	return err
}

func (cm *CMPair) handleClientReadEvent(data []byte) error {
	return cm.masterWrite(append([]byte{Input}, data...))
}

///webtty.go

// New creates a new instance of WeTTY.
// masterConn is a connection to the PTY master,
// typically it's a websocket connection to a client.
// slave is a PTY slave such as a local command with a PTY.
func NewMSPair(master Master, slave Slave) *MSPair {
	return &MSPair{
		master:     master,
		slave:      slave,
		bufferSize: 1024,
	}
}

// Run starts the main process of the WeTTY.
// This method blocks until the context is canceled.
// Note that the master and slave are left intact even
// after the context is canceled. Closing them is caller's
// responsibility.
// If the connection to one end gets closed, returns ErrSlaveClosed or ErrMasterClosed.
func (ms *MSPair) Pipe() error {
	errs := make(chan error, 2)

	// slave => buffer => master
	go func() {
		errs <- func() error {
			buffer := make([]byte, ms.bufferSize)
			for {
				n, err := ms.slave.Read(buffer)
				if err != nil {
					return ErrSlaveClosed
				}

				err = ms.handleSlaveReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	// slave <= buffer <= master
	go func() {
		errs <- func() error {
			buffer := make([]byte, ms.bufferSize)
			for {
				n, err := ms.master.Read(buffer)
				if err != nil {
					return ErrMasterClosed
				}

				err = ms.handleMasterReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	return <-errs
}

func (ms *MSPair) handleSlaveReadEvent(data []byte) error {
	return ms.masterWrite(append([]byte{Output}, data...))
}

func (ms *MSPair) masterWrite(data []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	_, err := ms.master.Write(data)
	return err
}

func (ms *MSPair) slaveWrite(data []byte) error {
	ms.mu2.Lock()
	defer ms.mu2.Unlock()
	_, err := ms.slave.Write(data)
	return err
}

func (ms *MSPair) handleMasterReadEvent(data []byte) error {
	if len(data) == 0 {
		return errors.New("unexpected zero length read from master")
	}

	switch data[0] {
	case Input:
		if len(data) <= 1 {
			return nil
		}
		if err := ms.slaveWrite(data[1:]); err != nil {
			return err //ors.Wrapf(err, "failed to write received data to slave")
		}
	case Ping:
		if err := ms.masterWrite([]byte{Pong}); err != nil {
			return err //ors.Wrapf(err, "failed to return Pong message to master")
		}
	case ResizeTerminal:
		sz := &pty.Winsize{}
		if len(data) <= 1 {
			return errors.New("received malformed remote command for terminal resize: empty payload")
		}
		if err := json.Unmarshal(data[1:], sz); err != nil {
			return err //ors.Wrapf(err, "received malformed data for terminal resize")
		}
		ms.slave.ResizeTerminal(sz)
		log.Println("new sz:", sz)
	default:
		return errors.New(fmt.Sprintf("unknown message type `%c`", data[0]))
	}

	return nil
}

///protocol.go

// Protocols defines the name of this protocol,
// which is supposed to be used to the subprotocol of Websockt streams.
// https://stackoverflow.com/questions/37984320/why-doesnt-golang-allow-const-maps
var (
	Protocols = []string{
		"webtty",
	}
	Dialer = &websocket.Dialer{
		Subprotocols: Protocols,
	}
	Upgrader = &websocket.Upgrader{
		Subprotocols: Protocols,
	}
)

const (
	/* those message types are used when reading from master */
	// Unknown message type, maybe sent by a bug
	UnknownInput = '0'
	// User input typically from a keyboard
	Input = '1'
	// Ping to the server from master(browser)
	Ping = '2'
	// Notify the server that the browser size has been changed, the server then tries to resize the slave tty
	ResizeTerminal = '3'

	/* those message types are used when writing to master */
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

	// ErrClientClosed is returned when the client connection is closed.
	ErrClientClosed = errors.New("client closed")
)
