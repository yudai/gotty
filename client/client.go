package client

import (
	"os"

	"github.com/navigaid/gotty/utils"
	"github.com/navigaid/gotty/wetty"
)

type Client struct {
	factory func() (wetty.Master, error)
}

func New(addr string) *Client {
	factory := func() (wetty.Master, error) {
		conn, _, err := wetty.Dialer.Dial(addr, nil)
		return &utils.WsWrapper{conn}, err
	}
	return &Client{
		factory: factory,
	}
}

func (c *Client) Run() (err error) {
	var client wetty.Client
	var master wetty.Master
	client = &utils.ReadWriter{os.Stdin, os.Stdout}
	master, err = c.factory()
	if err != nil {
		return err
	}
	return wetty.NewCMPair(client, master).Pipe()
}
