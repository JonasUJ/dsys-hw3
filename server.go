package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/JonasUJ/dsys-hw3/chittychat"
	"github.com/JonasUJ/dsys-hw3/lamport"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
)

type Server struct {
	chittychat.UnimplementedChatServer
	clients []*chittychat.Chat_ConnectServer
	chMsgs  chan *chittychat.Message
	time    uint64
	pid     uint32
}

func (server *Server) GetTime() uint64 {
	return server.time
}

func (server *Server) GetPid() uint32 {
	return server.pid
}

func (s *Server) Connect(stream chittychat.Chat_ConnectServer) error {
	s.clients = append(s.clients, &stream)

	l.Println("new client connection")
	// Send message to client to let them sync time
	stream.Send(lamport.MakeMessage(s, fmt.Sprintf("Welcome to the %s server", *name)))

	for {
		msg, err := stream.Recv()
		if err != nil {
			// Client stream closed: Forget stream.
			index := slices.Index(s.clients, &stream)
			s.clients = slices.Delete(s.clients, index, index+1)

			if err == io.EOF {
				// Connection closed gracefully
				s.chMsgs <- lamport.MakeMessage(s, "A client left the server")
				return nil
			} else {
				l.Printf("server recv err: %v\n", err)
				return err
			}
		}

		l.Printf("recv '%s'\n", msg)
		s.time = lamport.LamportRecv(s, msg)

		s.chMsgs <- msg
	}
}

func server() {
	// We need a listener for grpc
	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", *port))
	if err != nil {
		l.Fatalf("fail to listen on port %s: %v", *port, err)
	}
	defer listener.Close()

	server := &Server{
		clients: make([]*chittychat.Chat_ConnectServer, 0),
		chMsgs:  make(chan *chittychat.Message),
		pid: uint32(os.Getpid()),
	}

	go func() {
		grpcServer := grpc.NewServer()
		chittychat.RegisterChatServer(grpcServer, server)
		l.Printf("server %s is running on port %s", *name, *port)
		if err := grpcServer.Serve(listener); err != nil {
			l.Fatalf("stopped serving: %v", err)
		}
	}()

	// Recv messages and send them to everyone
	for msg := range server.chMsgs {
		l.Printf("send '%s'", msg)
		server.time = lamport.LamportSend(server)
		for _, client := range server.clients {
			(*client).Send(msg)
		}
	}
}
