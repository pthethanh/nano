package main

import (
	"context"

	"time"

	"github.com/google/uuid"
	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/grpc/health"
	"github.com/pthethanh/nano/grpc/interceptor/logging"
	"github.com/pthethanh/nano/grpc/interceptor/recovery"
	"github.com/pthethanh/nano/grpc/server"
	"github.com/pthethanh/nano/log"
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

func metricsInterceptor(metricSrv *memory.Reporter) grpc.UnaryServerInterceptor {
	c := metricSrv.Counter("grpc_requests_total", "method")
	h := metricSrv.Histogram("grpc_request_duration_seconds", []float64{0.1, 0.5, 1.0, 2.5, 5.0}, "method")
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		c.With("method", info.FullMethod).Add(1)
		start := time.Now()
		defer func() {
			h.With("method", info.FullMethod).Record(time.Since(start).Seconds())
		}()
		return handler(ctx, req)
	}
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
			server.ContextUnaryInterceptor(logging.ServerContextLogger(log.AppendToContext, map[string]func() string{
				"x-request-id": uuid.NewString,
			})),
			logging.UnaryServerInterceptor(log.Default(), logging.LogMethod(log.AppendToContext), logging.LogRequest(true), logging.LogReply(true)),
			grpc.UnaryServerInterceptor(metricsInterceptor(metricSrv)),
		),
		server.OnShutdown(func() { log.Info("cleaning up & saying goodbye....") }),
	)
	services := []any{new(HelloServer), newHealthServer(), metricSrv}
	if err := srv.ListenAndServe(context.TODO(), services...); err != nil {
		panic(err)
	}
}
