package kafka_test

import (
	"context"
	"sync"
	"testing"

	"github.com/IBM/sarama"
	"github.com/pthethanh/nano/plugins/broker/kafka"
)

// TestOpen_ConcurrentCallsDoNotRace proves (under `go test -race`) that
// concurrent Open() calls don't race on the connected-state guard. It uses
// an in-process sarama MockBroker so it doesn't need a real Kafka cluster.
func TestOpen_ConcurrentCallsDoNotRace(t *testing.T) {
	mb := sarama.NewMockBroker(t, 1)
	defer mb.Close()
	mb.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(mb.Addr(), mb.BrokerID()).
			SetController(mb.BrokerID()),
	})

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true

	b := kafka.New[string](
		kafka.Address[string](mb.Addr()),
		kafka.Config[string](cfg),
	)
	defer b.Close(context.Background())

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Open(context.Background())
		}()
	}
	wg.Wait()
}
