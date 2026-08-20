package server_test

import (
	"context"
	"testing"
	"time"

	server "github.com/pthethanh/nano/grpc/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// echoService implements the server service duck type with one hand-written
// unary method, so this test does not need generated protobuf service code.
type echoService struct{}

func (echoService) Register(s *grpc.Server) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "test.Echo",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Echo",
				Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
					in := new(wrapperspb.StringValue)
					if err := dec(in); err != nil {
						return nil, err
					}
					handler := func(ctx context.Context, req any) (any, error) {
						return wrapperspb.String("echo:" + req.(*wrapperspb.StringValue).Value), nil
					}
					if interceptor == nil {
						return handler(ctx, in)
					}
					return interceptor(ctx, in, &grpc.UnaryServerInfo{FullMethod: "/test.Echo/Echo"}, handler)
				},
			},
		},
	}, echoService{})
}

func TestNewTestRegistersServiceAndServesOverBufconn(t *testing.T) {
	conn := server.NewTest(t, []any{echoService{}})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := new(wrapperspb.StringValue)
	if err := conn.Invoke(ctx, "/test.Echo/Echo", wrapperspb.String("hi"), out); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if want := "echo:hi"; out.Value != want {
		t.Errorf("got %q, want %q", out.Value, want)
	}
}

func TestNewTestUnregisteredMethodReturnsUnimplemented(t *testing.T) {
	conn := server.NewTest(t, []any{echoService{}})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := new(wrapperspb.StringValue)
	err := conn.Invoke(ctx, "/test.DoesNotExist/Method", wrapperspb.String("hi"), out)
	if status.Code(err) != codes.Unimplemented {
		t.Errorf("got code=%v, want %v", status.Code(err), codes.Unimplemented)
	}
}
