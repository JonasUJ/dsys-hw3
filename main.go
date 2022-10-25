package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	start = flag.String("start", "client", "Entrypoint for the application. Either client or server")
	name  = flag.String("name", "NoName", "Name of this instance")
	port  = flag.String("port", "50050", "Port to connect to")
	l     = log.New(os.Stdout, "", log.Ltime)
)

func main() {
	flag.Parse()

	logfile := fmt.Sprintf("%s.log", *name)
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		l.Fatalf("could not open file %s: %v", logfile, err)
	}

	// Can't have two main functions in the same package
	if *start == "server" {
		l = log.New(io.MultiWriter(os.Stdout, file), "", log.Ltime)
		server()
	} else if *start == "client" {
		l.SetOutput(file)
		client()
	} else {
		l.Fatalf("start not a valid value '%s' - expected one of 'client' or 'server'", *start)
	}
}
