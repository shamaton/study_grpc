package main

import (
	"log"
	"net"

	pb "github.com/shamaton/study_grpc/eureka/proto"
	"github.com/shamaton/study_grpc/eureka/service"
	"github.com/shamaton/study_grpc/msgpack/encoding"
	"google.golang.org/grpc"
)

func main() {
	listenPort, err := net.Listen("tcp", ":19003")
	if err != nil {
		log.Fatalln(err)
	}

	server := grpc.NewServer(grpc.CustomCodec(encoding.New()))
	catService := &service.MyCatService{}
	// 実行したい実処理をseverに登録する
	pb.RegisterCatServer(server, catService)
	server.Serve(listenPort)
}