package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/grpc/health"
	"github.com/pthethanh/nano/grpc/server"
	"github.com/pthethanh/nano/log"
	"google.golang.org/grpc/metadata"
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

func requestIDLogger(ctx context.Context) (context.Context, error) {
	if ids := metadata.ValueFromIncomingContext(ctx, "x-request-id"); len(ids) > 0 {
		return log.AppendToContext(ctx, "x-request-id", ids[0]), nil
	}
	return log.AppendToContext(ctx, "x-request-id", uuid.NewString()), nil
}

func main() {
	srv := server.New(
		server.Address(":8081"),
		server.ChainUnaryInterceptor(server.ContextUnaryInterceptor(requestIDLogger)),
	)
	healthSrv := health.NewServer()
	healthSrv.AddService("hello", health.NoDelay, 5*time.Second, time.Second, health.CheckFunc(func(ctx context.Context) error {
		// ok
		return nil
	}))
	if err := srv.ListenAndServe(context.TODO(), new(HelloServer), healthSrv); err != nil {
		panic(err)
	}
}
