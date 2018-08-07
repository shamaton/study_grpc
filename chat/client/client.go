package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"

	pb "github.com/shamaton/study_grpc_go/chat/proto"
)

func main() {
	conn, err := grpc.Dial(":5000", grpc.WithInsecure())
	if err != nil {
		log.Fatalln("net.Dial:", err)
	}
	defer conn.Close()
	client := pb.NewChatClient(conn)

	var name string
	for {
		fmt.Print("name> ")
		if n, err := fmt.Scanln(&name); err == io.EOF {
			return
		} else if n == 0 {
			fmt.Println("name must be not empty")
			continue
		} else if n > 20 {
			fmt.Println("name must be less than or equal 20 characters")
			continue
		}
		break
	}

	sid, err := Authorize(client, name)
	if err != nil {
		log.Fatalln("authorize:", err)
	}

	events, err := Connect(client, sid)
	if err != nil {
		log.Fatalln("connect:", err)
	}

	go func() {
		for {
			select {
			case event := <-events:
				switch {
				case event.GetJoin() != nil:
					fmt.Printf("%s has joined.\n", event.GetJoin().Name)
				case event.GetLeave() != nil:
					fmt.Printf("%s has left.\n", event.GetJoin().Name)
				case event.GetLog() != nil:
					l := event.GetLog()
					fmt.Printf("%s> %s\n", l.Name, l.Message)
				}
			}
		}
	}()

	var message string
	for {
		fmt.Print("> ")
		if n, err := fmt.Scanln(&message); err == io.EOF {
			return
		} else if n > 0 {
			err := Say(client, sid, message)
			if err != nil {
				log.Fatalln("say:", err)
			}
		}
	}
}

func Authorize(client pb.ChatClient, name string) (sid []byte, err error) {
	req := pb.RequestAuthorize{
		Name: name,
	}
	res, err := client.Authorize(context.Background(), &req)
	if err != nil {
		return
	}
	sid = res.SessionId
	return
}

func Connect(client pb.ChatClient, sid []byte) (events chan *pb.Event, err error) {
	req := pb.RequestConnect{
		SessionId: sid,
	}
	stream, err := client.Connect(context.Background(), &req)
	if err != nil {
		return
	}
	events = make(chan *pb.Event, 1000)
	go func() {
		defer func() { close(events) }()
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalln("stream.Recv", err)
			}
			events <- event
		}
	}()
	return
}

func Say(client pb.ChatClient, sid []byte, message string) error {
	req := pb.CommandSay{
		SessionId: sid,
		Message:   message,
	}
	_, err := client.Say(context.Background(), &req)
	return err
}
