package main

import (
	"log"
	"os"
	"strings"

	"github.com/yudai/gotty/server"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalln("usage: gotty [command] [args]...")
	}

	srv, err := server.New(args)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("GoTTY is starting with command: %s", strings.Join(args, " "))
	log.Fatalln(srv.Run())
}
