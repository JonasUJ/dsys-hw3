package main

import (
	"log"
	"net"

	"github.com/JonasUJ/dsys-hw3/chittychat"
	"google.golang.org/grpc"
)

type Server struct {
	chittychat.UnimplementedChatServer
	clients []*chittychat.Chat_ConnectServer
	chMsgs chan *chittychat.Message
}

func (s *Server) Connect(stream chittychat.Chat_ConnectServer) error {
	s.clients = append(s.clients, &stream)

	go func() {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("client recv err: %v\n", err)
			return
		}
		s.chMsgs <- msg
	}()

	return nil
}

func server() {
	// We need a listener for grpc
	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", *port))
	if err != nil {
		log.Fatalf("fail to listen on port %s: %v", *port, err)
	}
	defer listener.Close()

	server := &Server{
		clients: make([]*chittychat.Chat_ConnectServer, 0),
		chMsgs: make(chan *chittychat.Message),
	}

	grpcServer := grpc.NewServer()
	chittychat.RegisterChatServer(grpcServer, server)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("stopped serving: %v", err)
	}

	// Recv messages and send them to everyone
	for {
		msg := <-server.chMsgs
		for _, client := range server.clients {
			(*client).Send(msg)
		}
	}
}
