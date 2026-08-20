package main

import (
	"context"

	"github.com/pthethanh/nano/examples/validation/api"
	"github.com/pthethanh/nano/grpc/server"
	"github.com/pthethanh/nano/validator"
	"google.golang.org/grpc"
)

type registrationServer struct {
	api.UnimplementedRegistrationServiceServer
}

func (*registrationServer) Register(_ context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	return &api.RegisterResponse{
		Message: "registered " + req.GetEmail(),
	}, nil
}

func main() {
	srv := server.New(
		server.Address(":8082"),
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(validator.Default()),
		),
	)
	if err := srv.ListenAndServe(context.Background(), new(registrationServer)); err != nil {
		panic(err)
	}
}
