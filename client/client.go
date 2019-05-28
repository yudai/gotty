package client

import (
	"encoding/json"
	"net"
	"os"
	"sync"

	"github.com/kr/pty"

	"github.com/navigaid/gotty/utils"
	"github.com/navigaid/gotty/wetty"
)

type Client struct {
	mu   *sync.Mutex
	conn net.Conn
	addr string
}

func New(addr string) *Client {
	return &Client{
		addr: addr,
		mu:   &sync.Mutex{},
	}
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
	c.conn = conn

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

	c.mu.Lock()
	c.conn.Write(append([]byte{wetty.ResizeTerminal}, resize...))
	c.mu.Unlock()
	// defer exec.Command("reset").Run()

	var client wetty.Client = &utils.ReadWriter{os.Stdin, os.Stdout}
	var master wetty.Master = conn
	return wetty.NewCMPair(client, master).Pipe()
}
