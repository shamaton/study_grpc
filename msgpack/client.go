package main

import (
	"context"
	"fmt"
	"log"

	"github.com/shamaton/study_grpc/msgpack/encoding"
	pb "github.com/shamaton/study_grpc/msgpack/proto"
	"google.golang.org/grpc"
)

func main() {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(encoding.Name)),
	}

	conn, err := grpc.Dial("127.0.0.1:19003", opts...)
	if err != nil {
		log.Fatal("client connection error:", err)
	}
	defer conn.Close()
	client := pb.NewCatClient(conn)
	message := &pb.GetMyCatMessage{TargetCat: "tama"}
	res, err := client.GetMyCat(context.TODO(), message)
	fmt.Printf("result:%#v \n", res)
	fmt.Printf("error::%#v \n", err)

	message = &pb.GetMyCatMessage{TargetCat: "mike"}
	res, err = client.GetMyCat(context.TODO(), message)
	fmt.Printf("result:%#v \n", res)
	fmt.Printf("error::%#v \n", err)

	message = &pb.GetMyCatMessage{TargetCat: "john"}
	res, err = client.GetMyCat(context.TODO(), message)
	fmt.Printf("result:%#v \n", res)
	fmt.Printf("error::%#v \n", err)
}
