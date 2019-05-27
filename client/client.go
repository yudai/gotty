package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/kr/pty"

	"github.com/yudai/gotty/wetty"
)

type Client struct {
	addr string
}

func New(addr string) *Client {
	return &Client{addr}
}

func (c *Client) Run() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: \r\nHijack: true\r\n\r\n"))
	if err != nil {
		return err
	}

	// read/write conn
	_ = conn

	// get terminal size
	size, err := pty.GetsizeFull(os.Stdin)
	if err != nil {
		return err
	}
	type termSize struct {
		Columns uint16
		Rows    uint16
	}
	resize, err := json.Marshal(termSize{size.Cols, size.Rows})
	if err != nil {
		return err
	}

	// os.Stderr.Write(resize)

	conn.Write(append([]byte{wetty.ResizeTerminal}, resize...))
	// defer exec.Command("reset").Run()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if n == 0 {
			println("continue")
			continue
		}
		switch buf[0] {
		case wetty.Output:
			os.Stdout.Write(buf[1:n])
		default:
			io.WriteString(os.Stderr, fmt.Sprintf("unrecognized message type: `%b`", buf[0]))
		}
	}

	return nil
}
