package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"

	"github.com/yudai/gotty/app"
	"github.com/yudai/gotty/backends/ptycommand"
	"github.com/yudai/gotty/utils"
)

func main() {
	cmd := cli.NewApp()
	cmd.Name = "gotty"
	cmd.Version = app.Version
	cmd.Usage = "Share your terminal as a web application"
	cmd.HideHelp = true
	cli.AppHelpTemplate = helpTemplate

	appOptions := &app.Options{}
	if err := utils.ApplyDefaultValues(appOptions); err != nil {
		exit(err, 1)
	}
	backendOptions := &ptycommand.Options{}
	if err := utils.ApplyDefaultValues(backendOptions); err != nil {
		exit(err, 1)
	}

	cliFlags, flagMappings, err := utils.GenerateFlags(appOptions, backendOptions)
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
			msg := "Error: No command given."
			cli.ShowAppHelp(c)
			exit(fmt.Errorf(msg), 1)
		}

		configFile := c.String("config")
		_, err := os.Stat(utils.ExpandHomeDir(configFile))
		if configFile != "~/.gotty" || !os.IsNotExist(err) {
			if err := utils.ApplyConfigFile(configFile, appOptions, backendOptions); err != nil {
				exit(err, 2)
			}
		}

		utils.ApplyFlags(cliFlags, flagMappings, c, appOptions, backendOptions)

		appOptions.EnableBasicAuth = c.IsSet("credential")
		appOptions.EnableTLSClientAuth = c.IsSet("tls-ca-crt")

		if err := app.CheckConfig(appOptions); err != nil {
			exit(err, 6)
		}

		manager, err := ptycommand.NewCommandClientContextManager(c.Args(), backendOptions)
		if err != nil {
			exit(err, 3)
		}
		app, err := app.New(manager, appOptions)
		if err != nil {
			exit(err, 3)
		}

		registerSignals(app)

		err = app.Run()
		if err != nil {
			exit(err, 4)
		}
	}
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
