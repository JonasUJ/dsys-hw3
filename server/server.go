package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")

func main() {
	// setLog() //uncomment this line to log to a log.txt file instead of the console

	flag.Parse()
	fmt.Println(".: server is staaaaarting :.")

	go launchServer()

	for {
		time.Sleep(time.Second * 5)
	}
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	/* list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	} */

	/*
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
	*/
}
