package localcmd

import (
	"github.com/yudai/gotty/server"
)

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
	return New(factory.command, argv)
}
