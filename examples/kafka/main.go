package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/config"
	"github.com/pthethanh/nano/log"
	"github.com/pthethanh/nano/plugins/broker/kafka"
)

type (
	Data struct {
		UUID string    `json:"uuid"`
		Kind string    `json:"kind"`
		Data string    `json:"data"`
		TS   time.Time `json:"ts"`
	}
	Conf struct {
		Addresses []string
		User      string
		Password  string
	}
)

func main() {
	log := log.Named("kafka")
	conf := config.MustRead[Conf](context.Background(), config.WithPaths("config.local", "yaml", "."))
	br := kafka.New(
		kafka.Address[Data](conf.Addresses),
		kafka.SASL[Data](conf.User, conf.Password),
		kafka.Logger[Data](log),
	)
	if err := br.Open(context.Background()); err != nil {
		log.Error("failed to open kafka connection", "error", err)
		return
	}
	defer br.Close(context.Background())
	sub, err := br.Subscribe(context.Background(), "test", func(e broker.Event[Data]) error {
		log.Info("Received", "data", e)
		return nil
	})
	if err != nil {
		log.Error("failed to subscribe", "error", err)
		return
	}
	defer sub.Unsubscribe()

	for i := 0; i < 10; i++ {
		if err := br.Publish(context.Background(), "test", &Data{
			UUID: uuid.NewString(),
			Kind: "iot",
			Data: fmt.Sprintf(`{"value":%d}`, i),
			TS:   time.Now(),
		}); err != nil {
			log.Error("failed to publish message", "error", err)
		}
	}
}
