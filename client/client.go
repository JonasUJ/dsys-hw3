package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	chittyChat "github.com/JonasUJ/dsys-hw3/chittychat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clientsName = flag.String("name", "default", "Senders Name") //Perhaps this is the name?
var serverPort = flag.String("server", "5400", "Tcp server")     //Is the sender no?

var server chittyChat.ChatServiceClient //the server
var ServerConn *grpc.ClientConn         //the server connection
//https://pkg.go.dev/context

var lamportTimeCounter = int64(0) // usigned int since we do not need negative numbers for lamport time.

func main() {
	//parse flags and arguments
	flag.Parse()

	fmt.Println("------- CLIENT APP -------")

	//log to file instead of console
	//setLog()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	context := context.Background()

	//find out how to exit properly
	defer ServerConn.Close()

	//kattis ftw
	//ways of scanning: fmt with scanf and bufio with NewScanner or NewReader, since we do not need buffering then NewScanner
	//https://www.reddit.com/r/golang/comments/ba387a/how_to_choose_between_bufionewscanner_and/
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go SendMessage(context, scanner.Text())
	}
}

func ConnectToServer() {
	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	//https://pkg.go.dev/google.golang.org/grpc#DialOption
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//https://pkg.go.dev/google.golang.org/grpc#Dial
	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = chittyChat.NewChatServiceClient(conn)
	ServerConn = conn

	log.Println("the connection is: ", conn.GetState().String())
}

func Connect(context context.Context) {
	chat := chittyChat.Chat{Name: *clientsName, Sender: *serverPort}

	messageStream, err := server.Connect(context, &chat)
	if err != nil {
		//fatalF to ensure call to os.Exit(1)
		log.Fatalf("I died %v", err)
	}
	SendMessage(context, "user joined")

	//perhaps replace with a for loop?
	waitc := make(chan struct{})

	//extract into a mtehod?
	go func() {
		for {
			message, err := messageStream.Recv() //definetly not inpu
			if err == io.EOF {
				close(waitc)
				return
			}
			//error handling please
			if *serverPort != message.Sender {
				if lamportTimeCounter < message.LamportTimeCounter {
					lamportTimeCounter = message.LamportTimeCounter
				} else {
					lamportTimeCounter++
				}
				//do a print with sender time and messsage
			}
		}
	}()
	<-waitc
}

func Disconnect(context context.Context) {
	chat := chittyChat.Chat{Name: *clientsName, Sender: *serverPort}
	server.Disconnect(context, &chat)
	SendMessage(context, "user left")
	//do a print
}

func SendMessage(context context.Context, messageContent string) {
	lamportTimeCounter++
	chat := chittyChat.Chat{Name: *clientsName, Sender: *serverPort}

	messageStream, err := server.Connect(context, &chat)

	if err != nil {
		easyLog("sending message failed", err)
	}
	//Error handling?

	message := chittyChat.Message{Content: messageContent, Sender: *serverPort, Chat: &chat, LamportTimeCounter: lamportTimeCounter}
	messageStream.SendMsg(&message)
}

func easyLog(errorMessage string, err error) {
	log.Printf("errorMessage: %v , error: %v", errorMessage, err)
}
