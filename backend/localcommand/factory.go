package localcommand

import (
	"syscall"

	"github.com/yudai/gotty/server"
)

type Options struct {
	CloseSignal  int `hcl:"close_signal" flagName:"close-signal" flagSName:"" flagDescribe:"Signal sent to the command process when gotty close it (default: SIGHUP)" default:"1"`
	CloseTimeout int `hcl:"close_timeout" flagName:"close-timeout" flagSName:"" flagDescribe:"Time in seconds to force kill process after client is disconnected (default: -1)" default:"-1"`
}

type Factory struct {
	command string
	argv    []string
	options *Options
}

func NewFactory(command string, argv []string, options *Options) (*Factory, error) {
	return &Factory{
		command: command,
		argv:    argv,
		options: options,
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
	return New(
		factory.command,
		argv,
		WithCloseSignal(syscall.Signal(factory.options.CloseSignal)),
		WithCloseSignal(syscall.Signal(factory.options.CloseTimeout)),
	)
}
