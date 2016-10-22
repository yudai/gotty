// +build darwin dragonfly freebsd netbsd openbsd

package reuseport

import (
	"runtime"
	"syscall"
)

var reusePort = syscall.SO_REUSEPORT

func maxListenerBacklog() int {
	var (
		n   uint32
		err error
	)
	switch runtime.GOOS {
	case "darwin", "freebsd":
		n, err = syscall.SysctlUint32("kern.ipc.somaxconn")
	case "netbsd":
		// NOTE: NetBSD has no somaxconn-like kernel state so far
	case "openbsd":
		n, err = syscall.SysctlUint32("kern.somaxconn")
	}
	if n == 0 || err != nil {
		return syscall.SOMAXCONN
	}
	// FreeBSD stores the backlog in a uint16, as does Linux.
	// Assume the other BSDs do too. Truncate number to avoid wrapping.
	// See issue 5030.
	if n > 1<<16-1 {
		n = 1<<16 - 1
	}
	return int(n)
}
