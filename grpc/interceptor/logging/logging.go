package logging

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	logger interface {
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
	}

	AppendToContextFunc = func(ctx context.Context, attrs ...any) context.Context
)

// ServerContextLogger returns a function that extracts logging attributes from gRPC metadata
// and appends them to the context using the provided appendToContext function.
// If a metadata key is not present, it uses the default value provided in attrsWithDefault.
// This function can be used with server.ContextUnaryInterceptor or server.ContextStreamInterceptor
// to enrich the context with logging attributes for each request.
func ServerContextLogger(appendToContext AppendToContextFunc, attrsWithDefault map[string]func() string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		newCtx := ctx
		for key, def := range attrsWithDefault {
			if vs := metadata.ValueFromIncomingContext(ctx, key); len(vs) > 0 {
				newCtx = appendToContext(newCtx, key, getValue(vs))
				continue
			}
			newCtx = appendToContext(newCtx, key, def())
		}
		return newCtx, nil
	}
}

// UnaryClientInterceptor returns a gRPC unary client interceptor that logs requests and responses.
// It uses the provided logger to log messages at various stages of the request handling process.
// The behavior of the interceptor can be customized using options such as logging the method name,
// request, reply, and duration.
func UnaryClientInterceptor(logger logger, opts ...Option) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) (err error) {
		t := time.Now()
		o := newOpts(opts...)
		newCtx := ctx
		appendedContext := o.appendToContext != nil
		if o.logMethod && o.appendToContext != nil {
			newCtx = o.appendToContext(ctx, "grpc.method", method)
		}
		if o.logRequest {
			attrs := []any{"grpc.request", req}
			if o.logMethod && !appendedContext {
				attrs = append(attrs, "grpc.method", method)
			}
			logger.Log(newCtx, slog.LevelInfo, "Sent gRPC request", attrs...)
		}
		if o.logReply {
			defer func() {
				attrs := []any{"grpc.reply", reply}
				if err != nil {
					attrs = append(attrs, "grpc.error", err)
				}
				if o.logDuration {
					attrs = append(attrs, "grpc.duration", time.Since(t).String())
				}
				if o.logMethod && !appendedContext {
					attrs = append(attrs, "grpc.method", method)
				}
				logger.Log(newCtx, slog.LevelInfo, "Received gRPC reply", attrs...)
			}()
		}
		return invoker(ctx, method, req, reply, cc, callOpts...)
	}
}

func StreamClientInterceptor(logger logger, opts ...Option) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, callOpts ...grpc.CallOption) (grpc.ClientStream, error) {
		// TODO implement logging for streaming RPCs
		return streamer(ctx, desc, cc, method, callOpts...)
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that logs requests and responses.
// It uses the provided logger to log messages at various stages of the request handling process.
// The behavior of the interceptor can be customized using options such as logging the method name,
// request, reply, and duration.
func UnaryServerInterceptor(logger logger, opts ...Option) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res any, err error) {
		t := time.Now()
		o := newOpts(opts...)
		newCtx := ctx
		appendedContext := o.appendToContext != nil
		if o.logMethod && o.appendToContext != nil {
			newCtx = o.appendToContext(ctx, "grpc.method", info.FullMethod)
		}
		if o.logRequest {
			attrs := []any{"grpc.request", req}
			if o.logMethod && !appendedContext {
				attrs = append(attrs, "grpc.method", info.FullMethod)
			}
			logger.Log(newCtx, slog.LevelInfo, "Received gRPC request", attrs...)
		}
		if o.logReply {
			defer func() {
				attrs := []any{"grpc.reply", res}
				if err != nil {
					attrs = append(attrs, "grpc.error", err)
				}
				if o.logDuration {
					attrs = append(attrs, "grpc.duration", time.Since(t).String())
				}
				if o.logMethod && !appendedContext {
					attrs = append(attrs, "grpc.method", info.FullMethod)
				}
				logger.Log(newCtx, slog.LevelInfo, "Sent gRPC reply", attrs...)
			}()
		}
		res, err = handler(newCtx, req)
		return res, err
	}
}

func StreamServerInterceptor(logger logger, opts ...Option) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// TODO implement logging for streaming RPCs
		return handler(srv, ss)
	}
}

func getValue(vs []string) string {
	if len(vs) == 0 {
		return ""
	}
	if len(vs) == 1 {
		return vs[0]
	}
	return fmt.Sprintf("[%s]", strings.Join(vs, ","))
}
