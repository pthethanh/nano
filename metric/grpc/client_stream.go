package grpc

import (
	"io"
	"sync"
	"time"

	"github.com/pthethanh/nano/metric"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type clientStream struct {
	gogrpc.ClientStream
	method    string
	startedAt time.Time
	counter   metric.Counter
	histogram metric.Histogram
	once      sync.Once
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
		code := grpcCode(err).String()
		s.counter.With("method", s.method, "code", code, "kind", "stream").Add(1)
		s.histogram.With("method", s.method, "code", code, "kind", "stream").Record(time.Since(s.startedAt).Seconds())
	})
}
