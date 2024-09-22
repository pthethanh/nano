package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
	return &api.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

func main() {
	srv := server.New(server.Address(":8081"))
	go testServer(srv)
	if err := srv.ListenAndServe(context.TODO(), new(HelloServer)); err != nil {
		panic(err)
	}
}

// testServer continuos call the server using both gRPC & HTTP.
// this is to demonstrate that we can call the server using both protocols
// without setting up anything else.
func testServer(srv *server.Server) {
	for {
		time.Sleep(time.Second)
		// call hello server using gRPC
		client := api.MustNewHelloClient(context.TODO(), srv.Address(), srv.DialOpts()...)
		res, err := client.SayHello(context.TODO(), &api.HelloRequest{
			Name: "Jack",
		})
		if err != nil {
			log.Error("gRPC error", "error", err)
			continue
		}
		log.Info("gRPC response", "message", res.Message)

		// call hello server using HTTP
		rs, err := http.Post("http://"+srv.Address()+"/api/v1/hello", "application/json", strings.NewReader(`{"name":"Jack"}`))
		if err != nil || rs.StatusCode != http.StatusOK {
			log.Error("HTTP error", "status", rs.Status, "error", err)
			continue
		}
		hrs := api.HelloResponse{}
		if err := json.NewDecoder(rs.Body).Decode(&hrs); err != nil {
			log.Error("HTTP unmarshal failed", "error", err)
			continue
		}
		log.Info("HTTP response", "message", hrs.Message)
	}
}
