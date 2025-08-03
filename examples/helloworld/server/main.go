package main

import (
	"context"

	"time"

	"github.com/google/uuid"

	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/grpc/health"
	"github.com/pthethanh/nano/grpc/server"
	"github.com/pthethanh/nano/log"
	"github.com/pthethanh/nano/metric/memory"
	"google.golang.org/grpc"
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

func loggerInterceptor(ctx context.Context) (context.Context, error) {
	if ids := metadata.ValueFromIncomingContext(ctx, "x-request-id"); len(ids) > 0 {
		return log.AppendToContext(ctx, "x-request-id", ids[0]), nil
	}
	return log.AppendToContext(ctx, "x-request-id", uuid.NewString()), nil
}

func recoverInterceptor(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorContext(ctx, "recovered from panic", "error", err)
		}
	}()
}

func metricsInterceptor(metricSrv *memory.Reporter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		metricSrv.Counter("grpc_requests_total", "method").With("method", info.FullMethod).Add(1)
		start := time.Now()
		defer func() {
			metricSrv.Histogram("grpc_request_duration_seconds", []float64{0.1, 0.5, 1.0, 2.5, 5.0}, "method").With("method", info.FullMethod).Record(time.Since(start).Seconds())
		}()
		return handler(ctx, req)
	}
}

func newHealthServer() *health.Server {
	healthSrv := health.NewServer()
	healthSrv.AddService("hello", health.NoDelay, 5*time.Second, time.Second, health.CheckFunc(func(ctx context.Context) error {
		// ok
		return nil
	}))
	return healthSrv
}

func main() {
	// Following APIs are available on HTTP:
	// POST /api/v1/hello
	// GET /api/v1/metrics
	// GET /api/v1/health/check
	// GET /api/v1/health/list
	metricSrv := memory.New()
	srv := server.New(
		server.Address(":8081"),
		server.Logger(log.Default()),
		server.ChainUnaryInterceptor(
			server.ContextUnaryInterceptor(loggerInterceptor),
			server.DeferContextUnaryInterceptor(recoverInterceptor),
			grpc.UnaryServerInterceptor(metricsInterceptor(metricSrv)),
		),
		server.OnShutdown(func() { log.Info("cleaning up & saying goodbye....") }),
	)
	services := []any{new(HelloServer), newHealthServer(), metricSrv}
	if err := srv.ListenAndServe(context.TODO(), services...); err != nil {
		panic(err)
	}
}
