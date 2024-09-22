# nano

A simple Go tool kit for building microservices.

## Design Principles

1. Easy to use.
2. Compatible with Go, gRPC native libraries as much as possible.
3. Default ready to use but give options for flexibility
4. No cross dependencies between packages.

## Getting Started

Define proto:
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

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the hello
message HelloResponse {
  string message = 1;
}
```

Implement the service and start the server:
```go
package main

import (
	"context"

	"github.com/pthethanh/nano/examples/helloworld/api"
	"github.com/pthethanh/nano/grpc/server"
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
	if err := srv.ListenAndServe(context.TODO(), new(HelloServer)); err != nil {
		panic(err)
	}
}
```

Make a call using `client` package:
```go
client := api.NewHelloClient(client.MustDial(context.TODO(), ":8081"))
res, err := client.SayHello(context.TODO(), &api.HelloRequest{
    Name: "Jack",
})
```

Make a call via REST API:
```shell
curl -X POST http://localhost:8081/api/v1/hello -d '{"name": "Jack"}'
```
See full code at [examples](https://github.com/pthethanh/nano/tree/main/examples/helloworld)