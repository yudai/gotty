package client

import "net"

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
	_ = conn
	return nil
}
