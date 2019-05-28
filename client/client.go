package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/kr/pty"

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

	go c.PingLoop()

	go c.ReadInput()

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
		case wetty.Pong:
		case wetty.Output:
			os.Stdout.Write(buf[1:n])
		default:
			io.WriteString(os.Stderr, fmt.Sprintf("unrecognized message type: `%s`", string(buf[0])))
		}
	}

	return nil
}

func (c *Client) PingLoop() {
	for range time.Tick(time.Second) {
		c.mu.Lock()
		c.conn.Write([]byte{wetty.Ping})
		c.mu.Unlock()
	}
}

func (c *Client) ReadInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		c.mu.Lock()
		io.WriteString(c.conn, string(wetty.Input)+line+"\n")
		c.mu.Unlock()
	}
}
