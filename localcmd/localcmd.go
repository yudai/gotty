// Package localcommand provides an implementation of webtty.Slave
// that launches a local command with a PTY.
package localcmd

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/kr/pty"
	"github.com/yudai/gotty/server"
)

// Factory implements the server.Factory interface
type Factory struct {
	command string
	argv    []string
}

func NewFactory(command string, argv []string) (*Factory, error) {
	return &Factory{
		command: command,
		argv:    argv,
	}, nil
}

func (factory *Factory) Name() string {
	return "local command"
}

func (factory *Factory) New(params map[string][]string) (server.Slave, error) {
	argv := make([]string, len(factory.argv))
	copy(argv, factory.argv)
	if params["arg"] != nil && len(params["arg"]) > 0 {
		argv = append(argv, params["arg"]...)
	}
	return NewLc(factory.command, argv)
}

// Lc implements the server.Slave interface
type Lc struct {
	command   string
	argv      []string
	cmd       *exec.Cmd
	pty       *os.File
	ptyClosed chan struct{}
}

func NewLc(command string, argv []string) (*Lc, error) {
	cmd := exec.Command(command, argv...)

	pty, err := pty.Start(cmd)
	if err != nil {
		// todo close cmd?
		return nil, err // ors.Wrapf(err, "failed to start command `%s`", command)
	}
	ptyClosed := make(chan struct{})

	lcmd := &Lc{
		command:   command,
		argv:      argv,
		cmd:       cmd,
		pty:       pty,
		ptyClosed: ptyClosed,
	}

	// When the process is closed by the user,
	// close pty so that Read() on the pty breaks with an EOF.
	go func() {
		defer func() {
			lcmd.pty.Close()
			close(lcmd.ptyClosed)
		}()

		lcmd.cmd.Wait()
	}()

	return lcmd, nil
}

func (lcmd *Lc) Read(p []byte) (n int, err error) {
	return lcmd.pty.Read(p)
}

func (lcmd *Lc) Write(p []byte) (n int, err error) {
	return lcmd.pty.Write(p)
}

func (lcmd *Lc) Close() error {
	if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		lcmd.cmd.Process.Signal(syscall.SIGINT)
	}
	for {
		select {
		case <-lcmd.ptyClosed:
			return nil
		}
	}
}

func (lcmd *Lc) ResizeTerminal(width int, height int) error {
	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(height),
		uint16(width),
		0,
		0,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		lcmd.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
