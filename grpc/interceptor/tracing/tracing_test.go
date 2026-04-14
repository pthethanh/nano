package tracing_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	nanotracing "github.com/pthethanh/nano/grpc/interceptor/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryClientInterceptorInjectsTraceContext(t *testing.T) {
	tp := newTracerProvider()

	interceptor := nanotracing.UnaryClientInterceptor(
		nanotracing.WithTracerProvider(tp),
		nanotracing.WithPropagator(propagation.TraceContext{}),
	)
	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		if got := md.Get("Traceparent"); len(got) == 0 {
			t.Fatal("expected traceparent metadata")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	spans := tp.Ended()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if got := spans[0].name; got != "/svc/method" {
		t.Fatalf("got span name %q, want %q", got, "/svc/method")
	}
}

func TestUnaryServerInterceptorExtractsParentAndRecordsStatus(t *testing.T) {
	tp := newTracerProvider()
	propagator := propagation.TraceContext{}
	parentCtx := oteltrace.ContextWithSpanContext(context.Background(), oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    traceID(1),
		SpanID:     spanID(1),
		TraceFlags: oteltrace.FlagsSampled,
	}))
	parentMD := metadata.MD{}
	propagator.Inject(parentCtx, metadataCarrier(parentMD))

	interceptor := nanotracing.UnaryServerInterceptor(
		nanotracing.WithTracerProvider(tp),
		nanotracing.WithPropagator(propagator),
	)
	_, err := interceptor(metadata.NewIncomingContext(context.Background(), parentMD), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/method"}, func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(grpcCodes.Internal, "boom")
	})
	if status.Code(err) != grpcCodes.Internal {
		t.Fatalf("got err=%v, want internal", err)
	}

	spans := tp.Ended()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if got := spans[0].parent.SpanID(); got != spanID(1) {
		t.Fatalf("got parent span ID %v, want %v", got, spanID(1))
	}
	if spans[0].statusCode != codes.Error {
		t.Fatalf("got status %v, want error", spans[0].statusCode)
	}
}

func TestUnaryServerInterceptorReturnsHandlerError(t *testing.T) {
	tp := newTracerProvider()
	interceptor := nanotracing.UnaryServerInterceptor(nanotracing.WithTracerProvider(tp))
	want := errors.New("boom")

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/method"}, func(ctx context.Context, req any) (any, error) {
		return nil, want
	})
	if !errors.Is(err, want) {
		t.Fatalf("got err=%v, want %v", err, want)
	}
}

func TestStreamClientInterceptorEndsSpanOnStreamCompletion(t *testing.T) {
	tp := newTracerProvider()
	interceptor := nanotracing.StreamClientInterceptor(
		nanotracing.WithTracerProvider(tp),
		nanotracing.WithPropagator(propagation.TraceContext{}),
	)

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{ClientStreams: true, ServerStreams: true}, nil, "/svc/method", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return &testClientStream{recvErr: io.EOF}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(tp.Ended()); got != 0 {
		t.Fatalf("got %d ended spans before stream completion, want 0", got)
	}

	if err := stream.RecvMsg(nil); err != io.EOF {
		t.Fatalf("got err=%v, want EOF", err)
	}

	spans := tp.Ended()
	if len(spans) != 1 {
		t.Fatalf("got %d ended spans, want 1", len(spans))
	}
	if spans[0].statusCode != codes.Ok {
		t.Fatalf("got status %v, want ok", spans[0].statusCode)
	}
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
	metadata.MD(c).Set(key, value)
}

func (c metadataCarrier) Keys() []string {
	md := metadata.MD(c)
	keys := make([]string, 0, len(md))
	for key := range md {
		keys = append(keys, key)
	}
	return keys
}

type tracerProvider struct {
	embedded.TracerProvider
	mu     sync.Mutex
	spans  []*recordingSpan
	nextID uint64
}

func newTracerProvider() *tracerProvider {
	return &tracerProvider{}
}

func (p *tracerProvider) Tracer(string, ...oteltrace.TracerOption) oteltrace.Tracer {
	return tracer{provider: p}
}

func (p *tracerProvider) Ended() []*recordingSpan {
	p.mu.Lock()
	defer p.mu.Unlock()
	ended := make([]*recordingSpan, 0, len(p.spans))
	for _, span := range p.spans {
		if span.ended {
			ended = append(ended, span)
		}
	}
	return ended
}

type tracer struct {
	embedded.Tracer
	provider *tracerProvider
}

func (t tracer) Start(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	cfg := oteltrace.NewSpanStartConfig(opts...)
	parent := oteltrace.SpanContextFromContext(ctx)

	t.provider.mu.Lock()
	t.provider.nextID++
	id := t.provider.nextID
	t.provider.mu.Unlock()

	traceID := parent.TraceID()
	if !traceID.IsValid() || cfg.NewRoot() {
		traceID = traceIDFor(id)
		parent = oteltrace.SpanContext{}
	}
	spanCtx := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID(id + 1),
		TraceFlags: oteltrace.FlagsSampled,
	})
	span := &recordingSpan{
		provider:    t.provider,
		name:        name,
		spanContext: spanCtx,
		parent:      parent,
		statusCode:  codes.Unset,
		attrs:       append([]attribute.KeyValue(nil), cfg.Attributes()...),
	}

	t.provider.mu.Lock()
	t.provider.spans = append(t.provider.spans, span)
	t.provider.mu.Unlock()

	return oteltrace.ContextWithSpan(ctx, span), span
}

type recordingSpan struct {
	embedded.Span
	provider    *tracerProvider
	name        string
	spanContext oteltrace.SpanContext
	parent      oteltrace.SpanContext
	statusCode  codes.Code
	statusDesc  string
	attrs       []attribute.KeyValue
	ended       bool
	recordedErr error
	mu          sync.Mutex
}

func (s *recordingSpan) End(...oteltrace.SpanEndOption) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ended = true
}

func (s *recordingSpan) AddEvent(string, ...oteltrace.EventOption) {}
func (s *recordingSpan) AddLink(oteltrace.Link)                    {}
func (s *recordingSpan) IsRecording() bool                         { return true }

func (s *recordingSpan) RecordError(err error, _ ...oteltrace.EventOption) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordedErr = err
}

func (s *recordingSpan) SpanContext() oteltrace.SpanContext { return s.spanContext }

func (s *recordingSpan) SetStatus(code codes.Code, description string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statusCode = code
	s.statusDesc = description
}

func (s *recordingSpan) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.name = name
}

func (s *recordingSpan) SetAttributes(kv ...attribute.KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attrs = append(s.attrs, kv...)
}

func (s *recordingSpan) TracerProvider() oteltrace.TracerProvider {
	return s.provider
}

type testClientStream struct {
	recvErr error
}

func (s *testClientStream) Header() (metadata.MD, error) { return nil, nil }
func (s *testClientStream) Trailer() metadata.MD         { return nil }
func (s *testClientStream) CloseSend() error             { return nil }
func (s *testClientStream) Context() context.Context     { return context.Background() }
func (s *testClientStream) SendMsg(any) error            { return nil }
func (s *testClientStream) RecvMsg(any) error            { return s.recvErr }

func traceID(n uint64) oteltrace.TraceID {
	var id oteltrace.TraceID
	id[15] = byte(n)
	return id
}

func traceIDFor(n uint64) oteltrace.TraceID {
	var id oteltrace.TraceID
	for i := range id {
		id[i] = byte(n + uint64(i))
	}
	return id
}

func spanID(n uint64) oteltrace.SpanID {
	var id oteltrace.SpanID
	id[7] = byte(n)
	return id
}
