package webtty

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"sync"
	"testing"
)

type mockMaster struct {
	gottyToMasterReader io.Reader
	gottyToMasterWriter io.Writer
	masterToGottyReader io.Reader
	masterToGottyWriter io.Writer
}

type mockSlave struct {
	gottyToSlaveReader io.Reader
	gottyToSlaveWriter io.Writer
	slaveToGottyReader io.Reader
	slaveToGottyWriter io.Writer
	columns, rows      int
}

func TestWriteFromPTY(t *testing.T) {
	mMaster := newMockMaster()
	mSlave := newMockSlave()

	dt, err := New(mMaster, mSlave)
	if err != nil {
		t.Fatalf("Unexpected error from New(): %s", err)
	}

	// Launch GoTTY
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		dt.Run(ctx)
	}()

	// Check that the initialization happens as expected
	checkNextMsgType(t, mMaster.gottyToMasterReader, SetWindowTitle)
	checkNextMsgType(t, mMaster.gottyToMasterReader, SetBufferSize)

	// Simulate the slave (the process being run by GoTTY)
	// echoing "foobar"
	buf := make([]byte, 1024)
	message := []byte("foobar")

	wg.Add(1)
	go func() {
		dt.handleSlaveReadEvent(message)
		wg.Done()
	}()

	// And then make sure it makes it way to the client
	// through the websocket as an output message
	n, err := mMaster.gottyToMasterReader.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected error from Read(): %s", err)
	}
	if buf[0] != Output {
		t.Fatalf("Unexpected message type `%c`", buf[0])
	}

	// Decode it and make sure it's intact
	decoded := make([]byte, 1024)
	n, err = base64.StdEncoding.Decode(decoded, buf[1:n])
	if err != nil {
		t.Fatalf("Unexpected error from Decode(): %s", err)
	}
	if !bytes.Equal(decoded[:n], message) {
		t.Fatalf("Unexpected message received: `%s`", decoded[:n])
	}

	cancel()
	wg.Wait()
}

func checkNextMsgType(t *testing.T, connInPipeReader io.Reader, expected byte) {
	msgType, _ := nextMsg(t, connInPipeReader)
	if msgType != expected {
		t.Fatalf("Unexpected message type `%c`", msgType)
	}
}

func nextMsg(t *testing.T, reader io.Reader) (byte, []byte) {
	buf := make([]byte, 1024)
	_, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	return buf[0], buf[1:]
}

func TestWriteFromConn(t *testing.T) {
	mMaster := newMockMaster()
	mSlave := newMockSlave()

	dt, err := New(mMaster, mSlave, WithPermitWrite())
	if err != nil {
		t.Fatalf("Unexpected error from New(): %s", err)
	}

	// Launch GoTTY
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		dt.Run(ctx)
	}()

	var (
		message []byte
		n       int
	)
	readBuf := make([]byte, 1024)

	// Absorb initialization messages
	mMaster.gottyToMasterReader.Read(readBuf)
	mMaster.gottyToMasterReader.Read(readBuf)

	// simulate input from frontend...
	message = []byte("1hello\n") // line buffered canonical mode
	wg.Add(1)
	go func() {
		dt.handleMasterReadEvent(message)
		wg.Done()
	}()

	// ...and make sure it makes it through to the slave intact
	n, err = mSlave.gottyToSlaveReader.Read(readBuf)
	if err != nil {
		t.Fatalf("Unexpected error from Write(): %s", err)
	}
	if !bytes.Equal(readBuf[:n], message[1:]) {
		t.Fatalf("Unexpected message received: `%s`", readBuf[:n])
	}

	// ping
	message = []byte("2\n") // line buffered canonical mode
	n, err = mMaster.masterToGottyWriter.Write(message)
	if err != nil {
		t.Fatalf("Unexpected error from Write(): %s", err)
	}
	if n != len(message) {
		t.Fatalf("Write() accepted `%d` for message `%s`", n, message)
	}

	n, err = mMaster.gottyToMasterReader.Read(readBuf)
	if err != nil {
		t.Fatalf("Unexpected error from Read(): %s", err)
	}
	if !bytes.Equal(readBuf[:n], []byte{'2'}) {
		t.Fatalf("Unexpected message received: `%s`", readBuf[:n])
	}

	// TODO: resize

	cancel()
	wg.Wait()
}

func newMockMaster() *mockMaster {
	rv := &mockMaster{}
	rv.gottyToMasterReader, rv.gottyToMasterWriter = io.Pipe()
	rv.masterToGottyReader, rv.masterToGottyWriter = io.Pipe()
	return rv
}

func (mm *mockMaster) Read(buf []byte) (int, error) {
	return mm.masterToGottyReader.Read(buf)
}

func (mm *mockMaster) Write(buf []byte) (int, error) {
	return mm.gottyToMasterWriter.Write(buf)
}

func newMockSlave() *mockSlave {
	rv := &mockSlave{}
	rv.gottyToSlaveReader, rv.gottyToSlaveWriter = io.Pipe()
	rv.slaveToGottyReader, rv.slaveToGottyWriter = io.Pipe()
	return rv
}

func (ms *mockSlave) Read(buf []byte) (int, error) {
	return ms.slaveToGottyReader.Read(buf)
}

func (ms *mockSlave) Write(buf []byte) (int, error) {
	return ms.gottyToSlaveWriter.Write(buf)
}

func (ms *mockSlave) WindowTitleVariables() map[string]interface{} {
	return nil
}

func (ms *mockSlave) ResizeTerminal(columns int, rows int) error {
	ms.columns = columns
	ms.rows = rows
	return nil
}
