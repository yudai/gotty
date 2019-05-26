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

	log.Printf("GoTTY is starting with command: %s", strings.Join(args, " "))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Fatalln(server.New(args).Run())
}
