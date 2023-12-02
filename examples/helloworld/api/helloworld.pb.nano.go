// Code generated by protoc-gen-nano. DO NOT EDIT.
// source: helloworld.proto

package api

import (
	fmt "fmt"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	proto "google.golang.org/protobuf/proto"
	math "math"
)

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (UnimplementedHelloServer) ServiceDesc() *grpc.ServiceDesc {
	return &Hello_ServiceDesc
}

func (UnimplementedHelloServer) RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) {
	RegisterHelloHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

func (UnimplementedHelloServer) Name() string {
	return "HelloServer"
}
