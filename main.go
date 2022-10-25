package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	start   = flag.String("start", "client", "Entrypoint for the application. Either client or server")
	name    = flag.String("name", "NoName", "Name of this instance")
	port    = flag.String("port", "50050", "Port to connect to")
)

func main() {
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime)

	logfile := fmt.Sprintf("%s.log", *name)
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("could not open file %s: %v", logfile, err)
	}
	log.SetOutput(file)

	// Can't have two main functions in the same package
	if *start == "server" {
		server()
	} else if *start == "client" {
		client()
	} else {
		log.Fatalf("start not a valid value '%s' - expected on of 'client' or 'server'", *start)
	}
}
