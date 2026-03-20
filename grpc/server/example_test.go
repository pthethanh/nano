package server_test

import (
	"net/http"
	"time"

	"github.com/pthethanh/nano/grpc/server"
)

func ExampleNew() {
	srv := server.New(
		server.SeparateAddresses(":8081", ":8080"),
		server.APIPrefix("/api"),
		server.GatewayForwardHeaders("X-Tenant-Id", "Authorization"),
		server.Timeout(5*time.Second, 10*time.Second),
	)
	_ = srv
}

func ExampleHandler() {
	srv := server.New(
		server.Address(":8081"),
		server.Handler("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})),
	)
	_ = srv
}
