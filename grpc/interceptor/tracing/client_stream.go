package tracing

import (
	"io"
	"sync"

	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type clientStream struct {
	grpc.ClientStream
	span oteltrace.Span
	once sync.Once
}

func (s *clientStream) Header() (metadata.MD, error) {
	header, err := s.ClientStream.Header()
	if err != nil {
		s.finish(err)
	}
	return header, err
}

func (s *clientStream) SendMsg(m any) error {
	err := s.ClientStream.SendMsg(m)
	if err != nil {
		s.finish(err)
	}
	return err
}

func (s *clientStream) RecvMsg(m any) error {
	err := s.ClientStream.RecvMsg(m)
	if err != nil {
		s.finish(err)
	}
	return err
}

func (s *clientStream) CloseSend() error {
	err := s.ClientStream.CloseSend()
	if err != nil {
		s.finish(err)
	}
	return err
}

func (s *clientStream) finish(err error) {
	s.once.Do(func() {
		if err == io.EOF {
			err = nil
		}
		record(s.span, err)
		s.span.End()
	})
}
