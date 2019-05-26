package main

import (
	"log"
	"os"
	"strings"

	"github.com/yudai/gotty/localcmd"
	"github.com/yudai/gotty/server"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalln("Error: No command given.")
	}

	factory, err := localcmd.NewFactory(args[0], args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	srv, err := server.New(factory)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("GoTTY is starting with command: %s", strings.Join(args, " "))
	log.Fatalln(srv.Run())
}
