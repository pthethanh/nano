package config

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type (
	Reader[T any] struct {
		opts *options
		vp   *viper.Viper
		log  Logger
	}

	Logger interface {
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
	}
)

// Read loads configuration into T using provided options.
func Read[T any](ctx context.Context, opts ...Option) (*T, error) {
	r, err := NewReader[T](opts...)
	if err != nil {
		return nil, err
	}
	t, err := r.Read(ctx)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// MustRead loads configuration and panics on error.
func MustRead[T any](ctx context.Context, opts ...Option) *T {
	t, err := Read[T](ctx, opts...)
	if err != nil {
		panic(err)
	}
	return t
}

// NewReader creates a new Reader for configuration loading.
func NewReader[T any](opts ...Option) (*Reader[T], error) {
	r := &Reader[T]{
		vp: viper.NewWithOptions(viper.ExperimentalBindStruct()),
		opts: &options{
			envReplacers: []*strings.Replacer{
				strings.NewReplacer(".", "_"),
			},
		},
		log: slog.Default(),
	}
	for _, apply := range opts {
		apply(r.opts)
	}

	if r.opts.envPrefix != "" {
		r.vp.SetEnvPrefix(r.opts.envPrefix)
	}
	for _, replacer := range r.opts.envReplacers {
		r.vp.SetEnvKeyReplacer(replacer)
	}
	if r.opts.local {
		r.vp.SetConfigName(r.opts.localName)
		r.vp.SetConfigType(r.opts.localType)
		// include current folder if not specify
		if len(r.opts.localPaths) == 0 {
			r.opts.localPaths = append(r.opts.localPaths, ".")
		}
		for _, p := range r.opts.localPaths {
			r.log.Log(context.Background(), slog.LevelInfo, "reading config", "path", filepath.Join(p, r.opts.localName+"."+r.opts.localType))
			r.vp.AddConfigPath(p)
		}
	}
	if r.opts.localFile {
		r.vp.SetConfigFile(r.opts.localFilePath)
	}
	if r.opts.remote {
		r.vp.SetConfigType(r.opts.remoteType)
		if err := r.vp.AddRemoteProvider(r.opts.remoteProvider, r.opts.remoteEndpoint, r.opts.remotePath); err != nil {
			return nil, err
		}
	}
	if r.opts.remoteSecured {
		r.vp.SetConfigType(r.opts.remoteType)
		if err := r.vp.AddSecureRemoteProvider(r.opts.remoteProvider, r.opts.remoteEndpoint, r.opts.remotePath, r.opts.remoteSecuredSecret); err != nil {
			return nil, err
		}
	}
	// always load env
	r.vp.AutomaticEnv()
	r.loadEnv()
	if r.opts.onChange != nil {
		r.vp.OnConfigChange(func(in fsnotify.Event) {
			r.opts.onChange(in)
		})
	}

	return r, nil
}

// MustNewReader creates a new Reader and panics on error.
func MustNewReader[T any](opts ...Option) *Reader[T] {
	r, err := NewReader[T](opts...)
	if err != nil {
		panic(err)
	}
	return r
}

// Read loads configuration into T from local or remote sources.
func (r *Reader[T]) Read(ctx context.Context) (*T, error) {
	if r.isLocal() {
		if err := r.vp.ReadInConfig(); err != nil {
			return nil, err
		}
	}
	if r.isRemote() {
		if err := r.vp.ReadRemoteConfig(); err != nil {
			return nil, err
		}
	}
	t := new(T)
	if err := r.vp.Unmarshal(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Reader[T]) isLocal() bool {
	return r.opts.local || r.opts.localFile
}

func (r *Reader[T]) isRemote() bool {
	return r.opts.remote || r.opts.remoteSecured
}

func (r *Reader[T]) loadEnv() {
	if len(r.opts.envFiles) > 0 {
		r.log.Log(context.Background(), slog.LevelInfo, "loading env file", "files", r.opts.envFiles)
		for _, p := range r.opts.envFiles {
			if _, err := os.Stat(p); err != nil {
				continue
			}
			if err := godotenv.Overload(p); err != nil {
				r.log.Log(context.Background(), slog.LevelError, "failed to load env file", "name", r.opts.envFiles, "error", err)
			}
		}
	}
}

// WriteEnv writes all config keys as environment variables to w.
func (r *Reader[T]) WriteEnv(w io.Writer) {
	for _, k := range r.vp.AllKeys() {
		env := k
		if r.vp.GetEnvPrefix() != "" {
			env = r.vp.GetEnvPrefix() + "." + k
		}
		for _, rp := range r.opts.envReplacers {
			env = rp.Replace(env)
		}
		env = strings.ToUpper(env) + "=" + r.vp.GetString(k)
		fmt.Fprintln(w, env)
	}
}
