package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/grpc/health"
	"github.com/pthethanh/nano/grpc/interceptor/logging"
	"github.com/pthethanh/nano/grpc/interceptor/recovery"
	"github.com/pthethanh/nano/grpc/interceptor/tracing"
	"github.com/pthethanh/nano/grpc/server"
	"github.com/pthethanh/nano/log"
	metricgrpc "github.com/pthethanh/nano/metric/grpc"
	"github.com/pthethanh/nano/metric/memory"
	"google.golang.org/grpc"
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

func newHealthServer() *health.Server {
	healthSrv := health.NewServer()
	healthSrv.Add(health.Service{
		Name:     "hello",
		Delay:    health.NoDelay,
		Timeout:  5 * time.Second,
		Interval: time.Second,
		Checker: health.CheckFunc(func(ctx context.Context) error {
			// check all dependencies status here
			return nil
		})})
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
		//server.SeparateAddresses(":8081", ":8082"),
		server.Logger(log.Default()),
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(),
			tracing.UnaryServerInterceptor(),
			server.ContextUnaryInterceptor(logging.ServerContextLogger(log.AppendToContext, map[string]any{
				"x-request-id": uuid.NewString,
			})),
			logging.UnaryServerInterceptor(log.Default(), logging.All()),
			metricgrpc.UnaryServerInterceptor(metricSrv),
		),
		server.OnShutdown(func() { log.Info("cleaning up & saying goodbye....") }),
	)
	services := []any{new(HelloServer), newHealthServer(), metricSrv}
	if err := srv.ListenAndServe(context.TODO(), services...); err != nil {
		panic(err)
	}
}
