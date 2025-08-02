package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/log"
	"github.com/pthethanh/nano/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	var serviceConfig = `{
		"loadBalancingConfig": [{"round_robin":{}}],
		"healthCheckConfig": {
			"serviceName": ""
		}
	}`
	hc := api.MustNewHelloClient(context.TODO(), srv, grpc.WithDefaultServiceConfig(serviceConfig))

	requestID := uuid.NewString()
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("X-Request-Id", requestID))
	ctx = log.AppendToContext(ctx, "X-Request-Id", requestID)
	log.InfoContext(ctx, "sending gRPC request")
	res, err := hc.SayHello(ctx, &api.HelloRequest{
		Name: "Jack",
	})
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "gRPC response", "message", res.Message)
	return nil
}

func sendHTTPRequest(srv string) error {
	url := "http://" + srv + "/api/v1/hello"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(`{"name":"Jack"}`))
	if err != nil {
		return err
	}
	requestID := uuid.NewString()
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Request-Id", requestID)
	req = req.WithContext(log.AppendToContext(req.Context(), "X-Request-Id", requestID))

	log.InfoContext(req.Context(), "sending HTTP request")
	rs, err := http.DefaultClient.Do(req)
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
	log.InfoContext(req.Context(), "HTTP response", "message", hrs.Message)
	return nil
}
