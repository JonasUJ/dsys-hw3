package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	gRPC "github.com/JonasUJ/dsys-hw3/chittychat"
	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChatServiceServer //FICXME: I suck at naming, what would be a good convetion or naming services?
	name                                string
	port                                string
	mutex                               sync.Mutex
}

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
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	/* 
		//If we want to make a server with options
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
	*/

	grpcServer := grpc.NewServer()

	server := &Server{
		name: *serverName,
		port: *port,
	
		gRPC.RegisterGetCurrentTimeServer(grpcServer, server) //Registers the server to the gRPC server.
	
		log.Printf("Server %s: Listening on port %s\n", *serverName, *port)
	
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to serve %v", err)
		}
	}

}

/* 
For an endpoint that does no streaming, then we need to give the method a context and the input type. For the return we need to return a pair of your return type and an error.
func (s *Server) <endpoint name>(ctx context.Context, <name> *<input type>) (*<the return type>, error) {
    //some code here
    ...
    ack :=  // make an instance of your return type
    return (ack, nil)
}
*/

/*
For an endpoint that streams messages, then we need to give the method a stream and return an error.
In this case you get the input from the stream and send the return type back over the stream too. 
func (s *Server) <endpoint name>(msgStream gRPC.<service name>_<endpoint name>Server) error {
    for {
        // get the next message from the stream
        msg, err := msgStream.Recv()
        if err == io.EOF {
            break
        }
    }


    ack := // make an instance of your return type
    msgStream.SendAndClose(ack)

    return nil
}
*/

// sets the logger to use a log.txt file instead of the console
func setLog() {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
}