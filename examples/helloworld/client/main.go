package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"strings"

	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/log"
	"github.com/pthethanh/nano/status"
)

func main() {
	var addr = flag.String("addr", ":8081", "server address")
	flag.Parse()

	if err := sendRPCRequest(*addr); err != nil {
		log.Error("failed to send gRPC request", "error", err)
	}
	if err := sendHTTPRequest(*addr); err != nil {
		log.Error("failed to send HTTP request", "error", err)
	}
}

func sendRPCRequest(srv string) error {
	client := api.MustNewHelloClient(context.TODO(), srv)
	res, err := client.SayHello(context.TODO(), &api.HelloRequest{
		Name: "Jack",
	})
	if err != nil {
		return err
	}
	log.Info("gRPC response", "message", res.Message)
	return nil
}

func sendHTTPRequest(srv string) error {
	rs, err := http.Post("http://"+srv+"/api/v1/hello", "application/json", strings.NewReader(`{"name":"Jack"}`))
	if err != nil {
		return err
	}
	if rs.StatusCode != http.StatusOK {
		return status.Internal("status_code: %v", rs.StatusCode)
	}
	var hrs api.HelloResponse
	if err := json.NewDecoder(rs.Body).Decode(&hrs); err != nil {
		return err
	}
	log.Info("HTTP response", "message", hrs.Message)
	return nil
}
