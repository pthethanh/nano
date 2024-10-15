package main

import (
	"context"

	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/grpc/server"
	"github.com/pthethanh/nano/log"
)

type (
	HelloServer struct {
		api.UnimplementedHelloServer
	}
)

func (*HelloServer) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloResponse, error) {
	log.InfoContext(ctx, "saying hello to", "name", req.Name)
	return &api.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

func main() {
	srv := server.New(server.Address(":8081"))
	if err := srv.ListenAndServe(context.TODO(), new(HelloServer)); err != nil {
		panic(err)
	}
}
