package ptycommand

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"syscall"
	"text/template"
	"unsafe"

	"github.com/yudai/gotty/backends"

	"github.com/kr/pty"
)

type Options struct {
	CloseSignal int    `hcl:"close_signal" flagName:"close-signal" flagSName:"" flagDescribe:"Signal sent to the command process when gotty close it (default: SIGHUP)" default:"1"`
	TitleFormat string `hcl:"title_format" flagName:"title-format" flagSName:"" flagDescribe:"Title format of browser window" default:"GoTTY - {{ .Command }} ({{ .Hostname }})"`
}

type CommandClientContextManager struct {
	command []string

	options       *Options
	titleTemplate *template.Template
}

func NewCommandClientContextManager(command []string, options *Options) (*CommandClientContextManager, error) {
	titleTemplate, err := template.New("title").Parse(options.TitleFormat)
	if err != nil {
		return nil, fmt.Errorf("Title format string syntax error: %v", options.TitleFormat)
	}
	return &CommandClientContextManager{command: command, options: options, titleTemplate: titleTemplate}, nil
}

type CommandClientContext struct {
	cmd *exec.Cmd
	pty *os.File
	mgr *CommandClientContextManager
}

func (mgr *CommandClientContextManager) New(params url.Values) (backends.ClientContext, error) {
	argv := mgr.command[1:]
	args := params["arg"]
	if len(args) != 0 {
		argv = append(argv, args...)
	}

	cmd := exec.Command(mgr.command[0], argv...)
	return &CommandClientContext{cmd: cmd, mgr: mgr}, nil
}

func (context *CommandClientContext) WindowTitle() (title string, err error) {
	hostname, _ := os.Hostname()

	titleVars := struct {
		Command  string
		Pid      int
		Hostname string
	}{
		Command:  context.cmd.Path,
		Pid:      context.cmd.Process.Pid,
		Hostname: hostname,
	}

	titleBuffer := new(bytes.Buffer)
	if err := context.mgr.titleTemplate.Execute(titleBuffer, titleVars); err != nil {
		return "", err
	}
	return titleBuffer.String(), nil
}

func (context *CommandClientContext) Start(exitCh chan bool) {
	ptyIo, err := pty.Start(context.cmd)
	if err != nil {
		log.Printf("failed to start command %v", err)
		exitCh <- true
	} else {
		context.pty = ptyIo
	}
}

func (context *CommandClientContext) InputWriter() io.Writer {
	return context.pty
}

func (context *CommandClientContext) OutputReader() io.Reader {
	return context.pty
}

func (context *CommandClientContext) ResizeTerminal(width, height uint16) error {
	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		height,
		width,
		0,
		0,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		context.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
	if errno != 0 {
		return errno
	} else {
		return nil
	}
}

func (context *CommandClientContext) TearDown() error {
	context.pty.Close()

	// Even if the PTY has been closed,
	// Read(0 in processSend() keeps blocking and the process doen't exit
	if context.cmd != nil && context.cmd.Process != nil {
		context.cmd.Process.Signal(syscall.Signal(context.mgr.options.CloseSignal))
		context.cmd.Wait()
	}
	return nil
}
