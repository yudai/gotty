package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"

	"github.com/yudai/gotty/app"
	"github.com/yudai/gotty/utils"
)

func main() {
	cmd := cli.NewApp()
	cmd.Version = app.Version
	cmd.Name = "gotty"
	cmd.Usage = "Share your terminal as a web application"
	cmd.HideHelp = true

	options := &app.Options{}
	if err := utils.ApplyDefaultValues(options); err != nil {
		exit(err, 1)
	}

	cliFlags, flagMappings, err := utils.GenerateFlags(options)
	if err != nil {
		exit(err, 3)
	}

	cmd.Flags = append(
		cliFlags,
		cli.StringFlag{
			Name:   "config",
			Value:  "~/.gotty",
			Usage:  "Config file path",
			EnvVar: "GOTTY_CONFIG",
		},
	)

	cmd.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			cli.ShowAppHelp(c)
			exit(fmt.Errorf("Error: No command given."), 1)
		}

		configFile := c.String("config")
		_, err := os.Stat(utils.ExpandHomeDir(configFile))
		if configFile != "~/.gotty" || !os.IsNotExist(err) {
			if err := utils.ApplyConfigFile(configFile, options); err != nil {
				exit(err, 2)
			}
		}

		utils.ApplyFlags(cliFlags, flagMappings, c, options)

		options.EnableBasicAuth = c.IsSet("credential")
		options.EnableTLSClientAuth = c.IsSet("tls-ca-crt")

		if err := app.CheckConfig(options); err != nil {
			exit(err, 6)
		}

		app, err := app.New(c.Args(), options)
		if err != nil {
			exit(err, 3)
		}

		registerSignals(app)

		err = app.Run()
		if err != nil {
			exit(err, 4)
		}
	}

	cli.AppHelpTemplate = helpTemplate

	cmd.Run(os.Args)
}

func exit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func registerSignals(app *app.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func() {
		for {
			s := <-sigChan
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				if app.Exit() {
					fmt.Println("Send ^C to force exit.")
				} else {
					os.Exit(5)
				}
			}
		}
	}()
}
