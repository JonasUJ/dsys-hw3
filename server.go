package main

import (
	"io"
	"net"

	"github.com/JonasUJ/dsys-hw3/chittychat"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
)

type Server struct {
	chittychat.UnimplementedChatServer
	clients []*chittychat.Chat_ConnectServer
	chMsgs  chan *chittychat.Message
}

func (s *Server) Connect(stream chittychat.Chat_ConnectServer) error {
	s.clients = append(s.clients, &stream)

	for {
		msg, err := stream.Recv()
		if err != nil {
			index := slices.Index(s.clients, &stream)
			s.clients = slices.Delete(s.clients, index, index+1)

			if err == io.EOF {
				return nil
			} else {
				l.Printf("server recv err: %v\n", err)
				return err
			}
		}

		l.Printf("got msg '%s'\n", msg.Content)

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
		for _, client := range server.clients {
			l.Printf("sending '%s'", msg)
			(*client).Send(msg)
		}
	}
}
