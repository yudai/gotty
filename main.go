package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"

	"github.com/yudai/gotty/app"
)

func main() {
	cmd := cli.NewApp()
	cmd.Version = app.Version
	cmd.Name = "gotty"
	cmd.Usage = "Share your terminal as a web application"
	cmd.HideHelp = true

	flags := []flag{
		flag{"address", "a", "IP address to listen"},
		flag{"port", "p", "Port number to listen"},
		flag{"permit-write", "w", "Permit clients to write to the TTY (BE CAREFUL)"},
		flag{"credential", "c", "Credential for Basic Authentication (ex: user:pass, default disabled)"},
		flag{"random-url", "r", "Add a random string to the URL"},
		flag{"random-url-length", "", "Random URL length"},
		flag{"tls", "t", "Enable TLS/SSL"},
		flag{"tls-crt", "", "TLS/SSL certificate file path"},
		flag{"tls-key", "", "TLS/SSL key file path"},
		flag{"tls-ca-crt", "", "TLS/SSL CA certificate file for client certifications"},
		flag{"index", "", "Custom index.html file"},
		flag{"title-format", "", "Title format of browser window"},
		flag{"reconnect", "", "Enable reconnection"},
		flag{"reconnect-time", "", "Time to reconnect"},
		flag{"timeout", "", "Timeout seconds for waiting a client (0 to disable)"},
		flag{"max-connection", "", "Maximum connection to gotty, 0(default) means no limit"},
		flag{"once", "", "Accept only one client and exit on disconnection"},
		flag{"permit-arguments", "", "Permit clients to send command line arguments in URL (e.g. http://example.com:8080/?arg=AAA&arg=BBB)"},
		flag{"close-signal", "", "Signal sent to the command process when gotty close it (default: SIGHUP)"},
		flag{"width", "", "Static width of the screen, 0(default) means dynamically resize"},
		flag{"height", "", "Static height of the screen, 0(default) means dynamically resize"},
	}

	mappingHint := map[string]string{
		"index":      "IndexFile",
		"tls":        "EnableTLS",
		"tls-crt":    "TLSCrtFile",
		"tls-key":    "TLSKeyFile",
		"tls-ca-crt": "TLSCACrtFile",
		"random-url": "EnableRandomUrl",
		"reconnect":  "EnableReconnect",
	}

	cliFlags, err := generateFlags(flags, mappingHint)
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
			fmt.Println("Error: No command given.\n")
			cli.ShowAppHelp(c)
			exit(err, 1)
		}

		options := app.DefaultOptions

		configFile := c.String("config")
		_, err := os.Stat(app.ExpandHomeDir(configFile))
		if configFile != "~/.gotty" || !os.IsNotExist(err) {
			if err := app.ApplyConfigFile(&options, configFile); err != nil {
				exit(err, 2)
			}
		}

		applyFlags(&options, flags, mappingHint, c)

		if c.IsSet("credential") {
			options.EnableBasicAuth = true
		}
		if c.IsSet("tls-ca-crt") {
			options.EnableTLSClientAuth = true
		}

		if err := app.CheckConfig(&options); err != nil {
			exit(err, 6)
		}

		app, err := app.New(c.Args(), &options)
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
