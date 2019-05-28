// Package localcommand provides an implementation of webtty.Slave
// that launches a local command with a PTY.
package localcmd

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/kr/pty"
)

// Factory implements the server.Factory interface
type Factory struct {
	Args []string
}

func (factory *Factory) New() (*Lc, error) {
	args := make([]string, len(factory.Args))
	copy(args, factory.Args)
	return NewLc(args)
}

// Lc implements the server.Slave interface {io.ReadWriteCloser,ResizeTerminal()}
type Lc struct {
	cmd       *exec.Cmd
	pty       *os.File
	ptyClosed chan struct{}
}

func NewLc(args []string) (*Lc, error) {
	cmd := exec.Command(args[0], args[1:]...)

	pty, err := pty.Start(cmd)
	if err != nil {
		// todo close cmd?
		return nil, err // ors.Wrapf(err, "failed to start command `%s`", command)
	}
	ptyClosed := make(chan struct{})

	lcmd := &Lc{
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

func (lcmd *Lc) ResizeTerminal(sz *pty.Winsize) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		lcmd.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(sz)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
