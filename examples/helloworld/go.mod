module github.com/pthethanh/nano/examples/helloworld

go 1.21.0

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0
	github.com/pthethanh/nano v0.0.2
	github.com/pthethanh/nano/cmd/protoc-gen-nano v0.0.2
	google.golang.org/genproto/googleapis/api v0.0.0-20240711142825-46eb208f015d
	google.golang.org/grpc v1.65.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.uber.org/zap/exp v0.2.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240711142825-46eb208f015d // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/pthethanh/nano v0.0.2 => ./../../

replace github.com/pthethanh/nano/cmd/protoc-gen-nano v0.0.2 => ./../../cmd/protoc-gen-nano/
