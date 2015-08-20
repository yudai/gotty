package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

	"github.com/yudai/gotty/app"
)

func main() {
	cmd := cli.NewApp()
	cmd.Version = "0.0.2"
	cmd.Name = "gotty"
	cmd.Usage = "Share your terminal as a web application"
	cmd.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "addr, a",
			Value:  "",
			Usage:  "IP address to listen",
			EnvVar: "GOTTY_ADDR",
		},
		cli.StringFlag{
			Name:   "port, p",
			Value:  "8080",
			Usage:  "Port number to listen",
			EnvVar: "GOTTY_PORT",
		},
		cli.BoolFlag{
			Name:   "permit-write, w",
			Usage:  "Permit clients to write to the TTY (BE CAREFUL)",
			EnvVar: "GOTTY_PERMIT_WRITE",
		},
		cli.StringFlag{
			Name:   "credential, c",
			Usage:  "Credential for Basic Authentication (ex: user:pass)",
			EnvVar: "GOTTY_CREDENTIAL",
		},
	}
	cmd.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			fmt.Println("Error: No command given.\n")
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		app := app.New(c.String("addr"), c.String("port"), c.Bool("permit-write"), c.String("credential"), c.Args())
		err := app.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	cmd.HideHelp = true
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [options] <command> [<arguments...>]

VERSION:
   {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`

	cmd.Run(os.Args)
}
