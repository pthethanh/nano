package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pthethanh/nano/config"
)

type (
	conf struct {
		Server server `mapstructure:"server"`
	}

	server struct {
		Host         string        `mapstructure:"host"`
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"readTimeout"`
		WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	}
)

func TestReadConfigPath(t *testing.T) {
	os.Clearenv()
	r := config.MustNewReader[conf](config.WithPaths("local", "yaml", "testdata"))
	conf, err := r.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	srv := server{
		Host:         "localhost",
		Port:         8000,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	if !cmp.Equal(conf.Server, srv) {
		t.Errorf("got server=%v, want server=%v", conf.Server, srv)
	}

}

func TestReadConfigPathWithEnvOpts(t *testing.T) {
	os.Clearenv()
	r := config.MustNewReader[conf](
		config.WithPaths("local_env_opts", "yaml", "testdata"),
		config.WithEnv("NANO", ".", "_"))
	conf, err := r.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	srv := server{
		Host:         "localhost",
		Port:         8000,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	if !cmp.Equal(conf.Server, srv) {
		t.Errorf("got server=%v, want server=%v", conf.Server, srv)
	}
}

func TestReadConfigPathWithEnvFile(t *testing.T) {
	os.Clearenv()
	r := config.MustNewReader[conf](
		config.WithPaths("local_env_file", "yaml", "testdata"),
		config.WithEnvFile("testdata/local_env_file_1.env", "NANO", ".", "_"))
	conf, err := r.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	srv := server{
		Host:         "localhost",
		Port:         8000,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if !cmp.Equal(conf.Server, srv) {
		t.Errorf("got server=%v, want server=%v", conf.Server, srv)
	}
}

func TestReadConfigPathWithFile(t *testing.T) {
	os.Clearenv()
	r := config.MustNewReader[conf](config.WithFile("testdata/local.yaml"))
	conf, err := r.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	srv := server{
		Host:         "localhost",
		Port:         8000,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	if !cmp.Equal(conf.Server, srv) {
		t.Errorf("got server=%v, want server=%v", conf.Server, srv)
	}
}
