package main

import (
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/shamaton/study_grpc_go/greeter/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type service struct{}

func (s *service) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Print("Received Say.Hello request")
	rsp := new(pb.HelloReply)
	rsp.Message = "Hello " + req.Name + ". " + fmt.Sprint(time.Now().Unix())
	return rsp, nil
}

func main() {

	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalln(err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &service{})
	s.Serve(l)

}
