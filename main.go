package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

	"github.com/yudai/gotty/app"
)

func main() {
	cmd := cli.NewApp()
	cmd.Version = "0.0.3"
	cmd.Name = "gotty"
	cmd.Usage = "Share your terminal as a web application"
	cmd.HideHelp = true
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
		cli.BoolFlag{
			Name:   "random-url, r",
			Usage:  "Add a random string to the URL",
			EnvVar: "GOTTY_RANDOM_URL",
		},
		cli.StringFlag{
			Name:   "title-format",
			Value:  "GoTTY - {{ .Command }} ({{ .Hostname }})",
			Usage:  "Title format of browser window",
			EnvVar: "GOTTY_title-format",
		},
	}
	cmd.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			fmt.Println("Error: No command given.\n")
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		app, err := app.New(
			app.Options{
				c.String("addr"),
				c.String("port"),
				c.Bool("permit-write"),
				c.String("credential"),
				c.Bool("random-url"),
				c.String("title-format"),
				c.Args(),
			},
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		err = app.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
	}

	cli.AppHelpTemplate = helpTemplate

	cmd.Run(os.Args)
}
