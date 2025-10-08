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
// Default values can be static values or functions that return a value (any, string).
func ServerContextLogger(appendToContext AppendToContextFunc, attrsWithDefault map[string]any) func(ctx context.Context) (context.Context, error) {
	return contextLogger(appendToContext, metadata.ValueFromIncomingContext, attrsWithDefault)
}

// ClientContextLogger returns a function that extracts logging attributes from gRPC metadata
// and appends them to the context using the provided appendToContext function.
// If a metadata key is not present, it uses the default value provided in attrsWithDefault.
// This function can be used with client.ContextUnaryInterceptor or client.ContextStreamInterceptor
// to enrich the context with logging attributes for each request.
// Default values can be static values or functions that return a value (any, string).
func ClientContextLogger(appendToContext AppendToContextFunc, attrsWithDefault map[string]any) func(ctx context.Context) (context.Context, error) {
	return contextLogger(appendToContext, func(ctx context.Context, key string) []string {
		md, _ := metadata.FromOutgoingContext(ctx)
		return md[key]
	}, attrsWithDefault)
}

// UnaryClientInterceptor returns a gRPC unary client interceptor that logs requests and responses.
// It uses the provided logger to log messages at various stages of the request handling process.
// The behavior of the interceptor can be customized using options such as logging the method name,
// request, reply, and duration.
func UnaryClientInterceptor(logger logger, opts ...Option) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) (err error) {
		t := time.Now()
		o := newOpts(opts...)
		logRequest(logger, ctx, "sent grpc request", o, method, req)
		defer func() {
			logResponse(logger, ctx, "received grpc response", o, method, reply, err, time.Since(t))
		}()
		return invoker(ctx, method, req, reply, cc, callOpts...)
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor that logs stream creation.
// It uses the provided logger to log messages when a stream is created.
// The behavior of the interceptor can be customized using options such as logging the method name and duration.
func StreamClientInterceptor(logger logger, opts ...Option) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, callOpts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		t := time.Now()
		o := newOpts(opts...)
		logRequest(logger, ctx, "creating grpc client stream", o, method, nil)
		defer func() {
			logResponse(logger, ctx, "created grpc client stream", o, method, nil, err, time.Since(t))
		}()
		stream, err = streamer(ctx, desc, cc, method, callOpts...)
		return stream, err
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
		logRequest(logger, ctx, "received grpc request", o, info.FullMethod, req)
		defer func() {
			logResponse(logger, ctx, "sent grpc response", o, info.FullMethod, res, err, time.Since(t))
		}()
		res, err = handler(ctx, req)
		return res, err
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that logs stream creation.
// It uses the provided logger to log messages when a stream is created.
// The behavior of the interceptor can be customized using options such as logging the method name and duration.
func StreamServerInterceptor(logger logger, opts ...Option) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		t := time.Now()
		o := newOpts(opts...)
		ctx := ss.Context()
		logRequest(logger, ctx, "creating grpc server stream", o, info.FullMethod, nil)
		defer func() {
			logResponse(logger, ctx, "created grpc server stream", o, info.FullMethod, nil, err, time.Since(t))
		}()
		err = handler(srv, ss)
		return err
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

func contextLogger(appendToContext AppendToContextFunc, metaFunc func(ctx context.Context, key string) []string, attrsWithDefault map[string]any) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		newCtx := ctx
		for key, def := range attrsWithDefault {
			if vs := metaFunc(ctx, key); len(vs) > 0 {
				newCtx = appendToContext(newCtx, key, getValue(vs))
				continue
			}
			if def != nil {
				if defFunc, ok := def.(func() any); ok {
					newCtx = appendToContext(newCtx, key, defFunc())
				} else if defFunc, ok := def.(func() string); ok {
					newCtx = appendToContext(newCtx, key, defFunc())
				} else {
					newCtx = appendToContext(newCtx, key, def)
				}
			} else {
				newCtx = appendToContext(newCtx, key, "")
			}

		}
		return newCtx, nil
	}
}

// logRequest logs the initial request with the specified method and request data.
func logRequest(logger logger, ctx context.Context, msg string, o *options, method string, req any) {
	attrs := []any{}
	if o.logMethod {
		attrs = append(attrs, "grpc.method", method)
	}
	if o.logRequest {
		attrs = append(attrs, "grpc.request", req)
	}
	logger.Log(ctx, slog.LevelInfo, msg, attrs...)
}

// logResponse logs the response with the specified method, response data, error, and duration.
func logResponse(logger logger, ctx context.Context, msg string, o *options, method string, res any, err error, duration time.Duration) {
	attrs := []any{}
	if o.logMethod {
		attrs = append(attrs, "grpc.method", method)
	}
	attrs = append(attrs, "grpc.response", res)
	attrs = append(attrs, "grpc.error", err)
	if o.logDuration {
		attrs = append(attrs, "grpc.duration", duration.String())
	}
	logger.Log(ctx, slog.LevelInfo, msg, attrs...)
}
