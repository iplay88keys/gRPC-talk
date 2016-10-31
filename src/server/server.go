package main

import (
	"fmt"
	"grpc/chat"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type chatServer struct {
	msgs          chan *chat.Message
	mu            sync.Mutex
	clientStreams map[chat.Chat_StreamMessagesServer]struct{}
}

func (c *chatServer) StreamMessages(stream chat.Chat_StreamMessagesServer) error {
	c.mu.Lock()
	c.clientStreams[stream] = struct{}{}
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.clientStreams, stream)
	}()
	return c.listenToClient(stream)
}

func (c *chatServer) Start() {
	var msg *chat.Message
	for {
		msg = <-c.msgs
		c.mu.Lock()
		for clientStream, _ := range c.clientStreams {
			clientStream.Send(msg)
		}
		c.mu.Unlock()
	}
}

func (c *chatServer) listenToClient(stream chat.Chat_StreamMessagesServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		c.msgs <- msg
	}
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Server listening on: %s", lis.Addr().String())

	grpcServer := grpc.NewServer()

	msgs := make(chan *chat.Message, 100)
	chatServer := &chatServer{
		msgs:          msgs,
		clientStreams: make(map[chat.Chat_StreamMessagesServer]struct{}),
	}
	chat.RegisterChatServer(grpcServer, chatServer)
	go chatServer.Start()
	grpcServer.Serve(lis)
}
