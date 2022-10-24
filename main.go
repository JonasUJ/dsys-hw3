package main

import (
	"flag"
	"log"
	"os"
)

var (
	start   = flag.String("start", "client", "Either client or server")
	name    = flag.String("name", "<no name>", "Name of this participant")
	port    = flag.String("port", "50050", "Port to connect to")
	logfile = flag.String("logfile", "", "Log file")
)

func main() {
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime)

	// Can't have two main functions in the same package
	if *start == "server" {
		server()
	} else {
		client()
	}
}
