package main

import (
	"log"
	"os"
	"strings"

	"github.com/navigaid/gotty/client"
	"github.com/navigaid/gotty/server"
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}

	args := os.Args[1:]

	// client mode
	if strings.HasSuffix(exe, "client") {
		if err := client.New("ws://localhost:8080/ws").Run(); err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}

	// server mode
	if len(args) == 0 {
		log.Fatalln("usage: gotty [command] [args]...")
	}

	log.Printf("GoTTY is starting with command: %s", strings.Join(args, " "))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Fatalln(server.New(args).Run())
}
