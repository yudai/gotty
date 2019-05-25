package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/yudai/gotty/localcmd"
	"github.com/yudai/gotty/server"
	"github.com/yudai/gotty/utils"
)

func main() {
	app := cli.NewApp()
	app.Name = "gotty"
	app.Usage = "Share your terminal as a web application"
	app.HideHelp = true

	serverOptions := &server.Options{}
	if err := utils.ApplyDefaultValues(serverOptions); err != nil {
		exit(err, 1)
	}
	localcmdOptions := &localcmd.Options{}
	if err := utils.ApplyDefaultValues(localcmdOptions); err != nil {
		exit(err, 1)
	}

	cliFlags, flagMappings, err := utils.GenerateFlags(serverOptions, localcmdOptions)
	if err != nil {
		exit(err, 3)
	}

	app.Flags = append(
		cliFlags,
		cli.StringFlag{
			Name:   "config",
			Value:  "~/.gotty",
			Usage:  "Config file path",
			EnvVar: "GOTTY_CONFIG",
		},
	)

	app.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			msg := "Error: No command given."
			cli.ShowAppHelp(c)
			exit(fmt.Errorf(msg), 1)
		}

		configFile := c.String("config")
		_, err := os.Stat(utils.Expand(configFile))
		if configFile != "~/.gotty" || !os.IsNotExist(err) {
			if err := utils.ApplyConfigFile(configFile, serverOptions, localcmdOptions); err != nil {
				exit(err, 2)
			}
		}

		utils.ApplyFlags(cliFlags, flagMappings, c, serverOptions, localcmdOptions)

		args := c.Args()
		factory, err := localcmd.NewFactory(args[0], args[1:], localcmdOptions)
		if err != nil {
			exit(err, 3)
		}

		srv, err := server.New(factory, serverOptions)
		if err != nil {
			exit(err, 3)
		}

		log.Printf("GoTTY is starting with command: %s", strings.Join(args, " "))

		errs := make(chan error, 1)
		go func() {
			errs <- srv.Run()
		}()
		select {}
	}
	app.Run(os.Args)
}

func exit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}
