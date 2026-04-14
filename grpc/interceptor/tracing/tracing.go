package tracing

import (
	"context"
	"io"
	"net/textproto"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

type options struct {
	tracerProvider oteltrace.TracerProvider
	propagator     propagation.TextMapPropagator
	attrs          []attribute.KeyValue
}

// Option customizes tracing interceptor behavior.
type Option func(*options)

// WithTracerProvider overrides the tracer provider used by the interceptors.
func WithTracerProvider(provider oteltrace.TracerProvider) Option {
	return func(o *options) {
		o.tracerProvider = provider
	}
}

// WithPropagator overrides the propagator used for metadata extraction and injection.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(o *options) {
		o.propagator = propagator
	}
}

// WithAttributes appends static attributes to every span created by the interceptors.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(o *options) {
		o.attrs = append(o.attrs, attrs...)
	}
}

// UnaryClientInterceptor returns a client interceptor that creates an OpenTelemetry span
// and injects the current span context into outgoing gRPC metadata.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := newOptions(opts...)
	tracer := o.tracerProvider.Tracer("github.com/pthethanh/nano/grpc/interceptor/tracing")

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		ctx, span := tracer.Start(ctx, method,
			oteltrace.WithSpanKind(oteltrace.SpanKindClient),
			oteltrace.WithAttributes(append(defaultAttributes("grpc.client", method), o.attrs...)...),
		)
		defer span.End()

		ctx = inject(ctx, o.propagator)
		err := invoker(ctx, method, req, reply, cc, callOpts...)
		record(span, err)
		return err
	}
}

// StreamClientInterceptor returns a client interceptor that traces the client
// stream lifecycle.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	o := newOptions(opts...)
	tracer := o.tracerProvider.Tracer("github.com/pthethanh/nano/grpc/interceptor/tracing")

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, callOpts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, span := tracer.Start(ctx, method,
			oteltrace.WithSpanKind(oteltrace.SpanKindClient),
			oteltrace.WithAttributes(append(defaultAttributes("grpc.client.stream", method), o.attrs...)...),
		)
		ctx = inject(ctx, o.propagator)
		stream, err := streamer(ctx, desc, cc, method, callOpts...)
		if err != nil {
			record(span, err)
			span.End()
			return nil, err
		}
		return &clientStream{ClientStream: stream, span: span}, nil
	}
}

// UnaryServerInterceptor returns a server interceptor that extracts parent span context
// from incoming gRPC metadata and creates a server span for the request.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := newOptions(opts...)
	tracer := o.tracerProvider.Tracer("github.com/pthethanh/nano/grpc/interceptor/tracing")

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = extract(ctx, o.propagator)
		ctx, span := tracer.Start(ctx, info.FullMethod,
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			oteltrace.WithAttributes(append(defaultAttributes("grpc.server", info.FullMethod), o.attrs...)...),
		)
		defer span.End()

		resp, err := handler(ctx, req)
		record(span, err)
		return resp, err
	}
}

// StreamServerInterceptor returns a server interceptor that traces accepted gRPC streams.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := newOptions(opts...)
	tracer := o.tracerProvider.Tracer("github.com/pthethanh/nano/grpc/interceptor/tracing")

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := extract(stream.Context(), o.propagator)
		ctx, span := tracer.Start(ctx, info.FullMethod,
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			oteltrace.WithAttributes(append(defaultAttributes("grpc.server.stream", info.FullMethod), o.attrs...)...),
		)
		defer span.End()

		err := handler(srv, &serverStream{ServerStream: stream, ctx: ctx})
		record(span, err)
		return err
	}
}

func newOptions(opts ...Option) *options {
	o := &options{
		tracerProvider: otel.GetTracerProvider(),
		propagator:     otel.GetTextMapPropagator(),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func defaultAttributes(kind, method string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.grpc.kind", kind),
		attribute.String("rpc.method", method),
	}
}

func record(span oteltrace.Span, err error) {
	if err == nil || err == io.EOF {
		span.SetStatus(codes.Ok, "")
		span.SetAttributes(attribute.String("rpc.grpc.status_code", grpcstatus.Code(err).String()))
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	span.SetAttributes(attribute.String("rpc.grpc.status_code", grpcstatus.Code(err).String()))
}

func inject(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	if md == nil {
		md = metadata.MD{}
	}
	propagator.Inject(ctx, metadataCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

func extract(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	if md == nil {
		md = metadata.MD{}
	}
	return propagator.Extract(ctx, metadataCarrier(md.Copy()))
}

type metadataCarrier metadata.MD

func (c metadataCarrier) Get(key string) string {
	values := metadata.MD(c).Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c metadataCarrier) Set(key, value string) {
	metadata.MD(c).Set(textproto.CanonicalMIMEHeaderKey(key), value)
}

func (c metadataCarrier) Keys() []string {
	md := metadata.MD(c)
	keys := make([]string, 0, len(md))
	for key := range md {
		keys = append(keys, key)
	}
	return keys
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}
