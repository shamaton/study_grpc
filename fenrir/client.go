package main

import (
	"fmt"
	"io"
	"log"

	pb "github.com/shamaton/study_grpc_go/fenrir/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:13009", grpc.WithInsecure())
	if err != nil {
		log.Fatalln("Dial:", err)
	}
	defer conn.Close()
	c := pb.NewMailClient(conn)

	/*
	 * Send
	 */
	msg := &pb.Message{
		From:    "sender",
		To:      []string{"test@example.com"},
		Subject: "hello",
		Body: &pb.Message_Text{
			Text: "hello world",
		},
	}
	if _, err := c.Send(context.Background(), msg); err != nil {
		log.Fatalln("Send:", err)
	}

	/*
	 * Receive
	 */
	stream, err := c.Receive(context.Background(), &pb.Folder{})
	if err != nil {
		log.Fatalln("Receive:", err)
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("Receive:", err)
		}
		fmt.Println(msg.Subject)
	}
}
