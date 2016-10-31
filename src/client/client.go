package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"grpc/chat"
	"os"

	"google.golang.org/grpc"
)

var serverAddr = flag.String("server_addr", "127.0.0.1:1000", "Server address in format of host:port")

func receiveMessages(name string, stream chat.Chat_StreamMessagesClient) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		if msg.User != name {
			fmt.Printf("%s> %s\n", msg.User, msg.Message)
		}
	}
}

func watchInput(name string, stream chat.Chat_StreamMessagesClient, reader *bufio.Reader) error {
	for {
		msg, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		chatMessage := &chat.Message{
			User:    name,
			Message: string(msg),
		}
		stream.Send(chatMessage)
	}
}

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := chat.NewChatClient(conn)
	stream, err := client.StreamMessages(context.Background())
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	input, _, err := reader.ReadLine()
	if err != nil {
		panic(err)
	}
	name := string(input)

	go watchInput(name, stream, reader)
	receiveMessages(name, stream)
}
