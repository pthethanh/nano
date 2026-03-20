package config_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pthethanh/nano/config"
)

type exampleConfig struct {
	Server struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"server"`
}

func ExampleRead() {
	dir, err := os.MkdirTemp("", "nano-config-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(path, []byte("server:\n  address: :8080\n"), 0o644); err != nil {
		panic(err)
	}

	cfg, err := config.Read[exampleConfig](context.Background(), config.WithFile(path))
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Server.Address)
	// Output: :8080
}

func ExampleReader_WriteEnv() {
	dir, err := os.MkdirTemp("", "nano-config-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(path, []byte("server:\n  address: :8080\n"), 0o644); err != nil {
		panic(err)
	}

	r, err := config.NewReader[exampleConfig](config.WithFile(path), config.WithEnv("APP", ".", "_"))
	if err != nil {
		panic(err)
	}
	if _, err := r.Read(context.Background()); err != nil {
		panic(err)
	}

	var out bytes.Buffer
	r.WriteEnv(&out)
	fmt.Print(out.String())
	// Output: APP_SERVER_ADDRESS=:8080
}
