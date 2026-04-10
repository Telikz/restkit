package main

import (
	"context"
	"fmt"
	"log"
	"net"

	rk "github.com/reststore/restkit"
	rkgrpc "github.com/reststore/restkit/extra/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/reststore/restkit/examples/grpc/proto"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest,
) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: fmt.Sprintf("Hello, %s!", req.Name)}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	pb.RegisterGreeterServer(gs, &server{})
	go gs.Serve(lis)

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewGreeterClient(conn)

	api := rk.NewApi()
	api.WithVersion("1.0.0").WithTitle("gRPC Example").WithSwaggerUI()

	greeter := rkgrpc.GRPC("/hello", client,
		func(ctx context.Context, c pb.GreeterClient, req *pb.HelloRequest) (*pb.HelloReply, error) {
			return c.SayHello(ctx, req)
		},
	)

	api.AddEndpoint(greeter)

	log.Println("Server: http://localhost:8080")
	if err := api.Serve(":8080"); err != nil {
		log.Fatal(err)
	}
}
