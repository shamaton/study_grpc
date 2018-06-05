package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/shamaton/grpc_study/fenrir/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Service struct{}

func (*Service) Send(ctx context.Context, msg *pb.Message) (*pb.Result, error) {
	fmt.Println("From:", msg.From)
	fmt.Println("To:", msg.To)
	body := msg.GetBody()
	if body == nil {
		return nil, fmt.Errorf("missing body")
	}
	switch v := body.(type) {
	case *pb.Message_Text:
		fmt.Println("Text:", v.Text)
	case *pb.Message_Html:
		fmt.Println("HTML:", v.Html)
	}
	return &pb.Result{}, nil
}

func (*Service) Receive(folder *pb.Folder, stream pb.Mail_ReceiveServer) error {
	replies := []*pb.Message{
		&pb.Message{
			From:    "from1",
			To:      []string{"to1", "to2"},
			Subject: "subject1",
			Body: &pb.Message_Text{
				Text: "plain text",
			},
		},
		&pb.Message{
			From:    "from2",
			To:      []string{"to3", "to4"},
			Subject: "subject2",
			Body: &pb.Message_Html{
				Html: "<div>attributed text</div>",
			},
		},
	}
	for _, r := range replies {
		if err := stream.Send(r); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	l, err := net.Listen("tcp", ":13009")
	if err != nil {
		log.Fatalln(err)
	}
	s := grpc.NewServer()
	pb.RegisterMailServer(s, &Service{})
	s.Serve(l)
}
