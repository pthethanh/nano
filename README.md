# nano

`nano` is a modular Go toolkit for building microservices.

It is designed to work well both for human developers and for code agents consuming it as a dependency in another project.

## Design Principles

1. Simple and easy to use.
2. Prefer native Go and gRPC patterns over heavy abstractions.
3. Ready to use out of the box, with flexible configuration options.
4. No cross-package dependencies. Each top-level package is intended to be used independently.

## Using `nano` In Another Project

If you are using `nano` from another repository, import only the packages you need.

Common starting points:
- `github.com/pthethanh/nano/grpc/server`: gRPC server setup with optional HTTP gateway support
- `github.com/pthethanh/nano/grpc/client`: gRPC client helpers, metadata forwarding, and production dial options
- `github.com/pthethanh/nano/config`: config loading from file, env, and remote providers
- `github.com/pthethanh/nano/log`: context-aware structured logging built on `log/slog`
- `github.com/pthethanh/nano/status`: gRPC-compatible status helpers and HTTP mapping utilities
- `github.com/pthethanh/nano/validator`: request and struct validation helpers
- `github.com/pthethanh/nano/broker`: async message broker interface and implementations
- `github.com/pthethanh/nano/cache`: cache interface and implementations
- `github.com/pthethanh/nano/metric`: metric interfaces and in-memory reporter
- `github.com/pthethanh/nano/grpc/interceptor/...`: composable gRPC middleware for auth, recovery, tracing, retry, rate limiting, and circuit breaking
- `github.com/pthethanh/nano/metric/grpc`: reusable gRPC client/server metrics interceptors

Recommended adoption path:
1. Start with `grpc/server` and `grpc/client` if you are building a gRPC service.
2. Add `config` for configuration loading.
3. Add `log` for context-aware logging.
4. Use `status` for consistent service errors.
5. Add `validator`, `broker`, `cache`, or `metric` only where needed.

## Minimal gRPC Server

If you already have generated gRPC service stubs, `grpc/server` is the main entry point.

```go
package main

import (
	"context"

	"github.com/pthethanh/nano/grpc/server"
	"github.com/pthethanh/nano/log"
)

type HelloServer struct {
	api.UnimplementedHelloServer
}

func (*HelloServer) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloResponse, error) {
	log.InfoContext(ctx, "saying hello", "name", req.Name)
	return &api.HelloResponse{Message: "Hello " + req.Name}, nil
}

func main() {
	srv := server.New(server.Address(":8081"))
	if err := srv.ListenAndServe(context.Background(), new(HelloServer)); err != nil {
		panic(err)
	}
}
```

## Minimal gRPC Client

```go
package main

import (
	"context"

	nanoclient "github.com/pthethanh/nano/grpc/client"
)

func main() {
	conn := nanoclient.MustNew(context.Background(), ":8081")
	defer conn.Close()

	client := api.NewHelloClient(conn)
	res, err := client.SayHello(context.Background(), &api.HelloRequest{Name: "Jack"})
	if err != nil {
		panic(err)
	}
	_ = res
}
```

## Config Example

```go
package main

import (
	"context"

	"github.com/pthethanh/nano/config"
)

type Config struct {
	Server struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"server"`
}

func main() {
	cfg, err := config.Read[Config](
		context.Background(),
		config.WithFile("config.local.yaml"),
		config.WithEnv("APP", ".", "_"),
	)
	if err != nil {
		panic(err)
	}
	_ = cfg
}
```

## Logging Example

```go
package main

import (
	"context"

	"github.com/pthethanh/nano/log"
)

func main() {
	ctx := log.AppendToContext(context.Background(), "request_id", "req-123")
	log.InfoContext(ctx, "request started")
}
```

## Status Example

```go
func FindUser(id string) error {
	if id == "" {
		return status.InvalidArgument("id is required")
	}
	return status.NotFound("user %q not found", id)
}
```

## Code Generation

If you want `nano` to generate service scaffolding, install the generator:

```shell
go install github.com/pthethanh/nano/cmd/protoc-gen-nano
```

Example `protoc` command:

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

See the full generated example in [`examples/helloworld`](./examples/helloworld).

## Examples

- [`examples/helloworld`](./examples/helloworld): generated gRPC service, client, and HTTP gateway
- [`examples/kafka`](./examples/kafka): broker integration example

## Validation

Useful repository commands:

```shell
go test ./...
make test
make fmt
make vet
```

The root `Makefile` also builds plugin and example modules.

## Features

- **gRPC**: gRPC server and client helpers with HTTP gateway support
- **Interceptors**: tracing, retry, rate limiting, circuit breaking, recovery, and logging helpers
- **Broker**: asynchronous messaging interface with pluggable backends
- **Cache**: generic cache interface with in-memory and Redis implementations
- **Config**: configuration loading from files, env, and remote providers
- **Log**: structured, context-aware logging with `slog`
- **Metrics**: counters, gauges, histograms, summaries, and reusable gRPC interceptors
- **Status**: gRPC-compatible status helpers and HTTP mapping utilities
- **Plugins**: optional plugin modules for broker and cache backends
- **protoc-gen-nano**: code generator for service scaffolding and gateway output

If you are driving code generation with an agent, prefer these inputs:
- one or two working examples from `examples/`
- the package docs on `grpc/server`, `grpc/client`, `config`, `log`, and `status`
- the repository rules in [`AGENTS.md`](./AGENTS.md)
