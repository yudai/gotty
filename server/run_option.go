package server

import (
	"context"
	"net"
)

// RunOptions holds a set of configurations for Server.Run().
type RunOptions struct {
	gracefullCtx context.Context

	listener net.Listener
}

// RunOption is an option of Server.Run().
type RunOption func(*RunOptions)

// WithGracefullContext accepts a context to shutdown a Server
// with care for existing client connections.
func WithGracefullContext(ctx context.Context) RunOption {
	return func(options *RunOptions) {
		options.gracefullCtx = ctx
	}
}

// WithListener allows the caller to override the default listener
func WithListener(l net.Listener) RunOption {
	return func(options *RunOptions) {
		options.listener = l
	}
}
