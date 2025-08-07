package config_test

import (
	"bytes"
	"context"
	"os"
	"strings"
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

func TestMustNewReaderWithFile(t *testing.T) {
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

func TestReadConfig(t *testing.T) {
	os.Clearenv()
	conf, err := config.Read[conf](context.Background(), config.WithFile("testdata/local.yaml"))
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

func TestMustReadConfig(t *testing.T) {
	os.Clearenv()
	conf := config.MustRead[conf](context.Background(), config.WithFile("testdata/local.yaml"))
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

func TestWriteEnv(t *testing.T) {
	os.Clearenv()
	r, err := config.NewReader[conf](
		config.WithEnv("TEST"),
		config.WithFile("testdata/local.yaml"),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := bytes.NewBuffer(nil)
	_, _ = r.Read(context.Background())
	r.WriteEnv(got)
	want := `TEST_SERVER_PORT=8000`
	if !strings.Contains(got.String(), want) {
		t.Errorf("got env=%v, want env=%v", got.String(), want)
	}
}
