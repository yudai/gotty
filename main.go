package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

	"github.com/yudai/gotty/app"
)

func main() {
	cmd := cli.NewApp()
	cmd.Version = "0.0.1"
	cmd.Name = "gotty"
	cmd.Usage = "Share your terminal as a web application"
	cmd.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "addr, a",
			Value:  "",
			Usage:  "IP address to listen at",
			EnvVar: "GOTTY_ADDR",
		},
		cli.StringFlag{
			Name:   "port, p",
			Value:  "8080",
			Usage:  "Port number to listen at",
			EnvVar: "GOTTY_PORT",
		},
		cli.BoolFlag{
			Name:   "permit-write, w",
			Usage:  "Permit write from client (BE CAREFUL)",
			EnvVar: "GOTTY_PERMIT_WRITE",
		},
	}
	cmd.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			fmt.Println("No command  given.\n")
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		app := app.New(c.String("addr"), c.String("port"), c.Bool("permit-write"), c.Args())
		err := app.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}
	cmd.Run(os.Args)
}
