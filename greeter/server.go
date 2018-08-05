package main

import (
	"log"
	"net"

	pb "github.com/shamaton/study_grpc/greeter/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Service struct{}

func (s *Service) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Print("Received Say.Hello request")
	rsp := new(pb.HelloReply)
	rsp.Message = "Hello " + req.Name
	return rsp, nil
}

func main() {

	l, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln(err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &Service{})
	s.Serve(l)

	/*
		service := grpc.NewService(
			micro.Name("go.micro.srv.greeter"),
		)

		// optionally setup command line usage
		service.Init()

		// Register Handlers
		hello.RegisterSayHandler(service.Server(), new(Say))

		// Run server
		if err := service.Run(); err != nil {
			log.Fatal(err)
		}
	*/
}
