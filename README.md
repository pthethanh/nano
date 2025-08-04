# nano

A clean, simple, and easy-to-use Go toolkit for building microservices.

## Design Principles

1. Simple and intuitive APIs.
2. Seamless integration with Go and native gRPC libraries.
3. Ready to use out of the box, with flexible configuration.
4. No cross-package dependencies — each package is standalone.

## Getting Started

### 1. Define your proto

```proto
syntax = "proto3";

package helloworld;
option go_package = "github.com/pthethanh/nano/examples/helloworld/api;api";

import "google/api/annotations.proto";

// The hello service definition.
service Hello {
  // Sends a hello
  rpc SayHello (HelloRequest) returns (HelloResponse) {
    option (google.api.http) = {
      post: "/api/v1/hello"
      body: "*"
    };
  }
}

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
}
```

### 2. Install the nano code generator

```shell
go install github.com/pthethanh/nano/cmd/protoc-gen-nano
```

### 3. Generate code

```shell
protoc -I /usr/local/include -I ~/go/src/github.com/googleapis/googleapis -I ./proto \
 --go_out ~/go/src \
 --nano_out ~/go/src \
 --nano_opt generate_gateway=true \
 --go-grpc_out ~/go/src \
 --grpc-gateway_out ~/go/src \
 --grpc-gateway_opt logtostderr=true \
 --grpc-gateway_opt generate_unbound_methods=true \
 ./proto/helloworld.proto
```

### 4. Implement the service and start the server

```go
type HelloServer struct {
	api.UnimplementedHelloServer
}

func (*HelloServer) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloResponse, error) {
	log.InfoContext(ctx, "saying hello to", "name", req.Name)
	return &api.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

func main() {
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

```

### 5. Create a client and call the server

```go
client := api.MustNewHelloClient(context.TODO(), ":8081")
res, err := client.SayHello(context.TODO(), &api.HelloRequest{
    Name: "Jack",
})
```

### 6. Call via REST API

```shell
curl -X POST http://localhost:8081/api/v1/hello -d '{"name": "Jack"}'
```

---

See full code at [examples/helloworld](https://github.com/pthethanh/nano/tree/main/examples/helloworld)

## Features

- **gRPC**: Production-ready, user-friendly APIs for building gRPC servers and clients. Includes built-in support for interceptors and HTTP gateway integration.
- **Broker**: Asynchronous messaging interface with pluggable backends such as in-memory, Kafka, and NATS.
- **Cache**: Generic caching interface supporting multiple backends like in-memory and Redis, with plugin extensibility.
- **Config**: Versatile configuration loader that supports local files, environment variables, and remote providers.
- **Log**: Structured, context-aware logging with support for various backends.
- **Metrics**: Unified interface for collecting metrics—including counters, gauges, histograms, and summaries—compatible with Prometheus and other reporters.
- **Status**: Comprehensive error and status handling aligned with gRPC codes and HTTP status mappings.
- **Plugins**: Extend broker and cache functionality effortlessly using custom plugins.
- **protoc-gen-nano**: Built-in code generator for scaffolding gRPC services, HTTP gateways, and nano service templates.

All components are modular and can be used independently or together—making **nano** ideal for building scalable, maintainable microservices.