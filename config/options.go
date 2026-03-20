package config

import (
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type (
	options struct {
		// env
		env          bool
		envFiles     []string
		envPrefix    string
		envReplacers []*strings.Replacer

		// local name & paths
		local      bool
		localName  string
		localPaths []string
		localType  string

		// local file
		localFile     bool
		localFilePath string

		// remote
		remote         bool
		remoteType     string
		remoteProvider string
		remoteEndpoint string
		remotePath     string

		// secured remote
		remoteSecured       bool
		remoteSecuredSecret string

		// logger
		logger Logger

		onChange func(in fsnotify.Event)
	}

	// Option configures how Reader discovers and loads configuration.
	Option func(*options)
)

// WithPaths loads configuration by name and type from one or more search paths.
//
// For each provided path, Reader also attempts to load a sibling `<name>.env`
// file before reading the main config file.
func WithPaths(name string, typ string, paths ...string) Option {
	return func(opts *options) {
		opts.local = true
		opts.localName = name
		opts.localType = typ
		opts.localPaths = paths

		// set default env
		for _, p := range paths {
			opts.envFiles = append(opts.envFiles, filepath.Join(p, name+".env"))
		}
	}
}

// WithFile loads configuration from a specific file path.
//
// Unless env loading was already configured explicitly, Reader also attempts to
// load `<base>.env` from the same directory, where `<base>` is the file name
// without its extension.
func WithFile(file string) Option {
	return func(opts *options) {
		opts.localFile = true
		opts.localFilePath = file

		// set default env
		if !opts.env {
			name := filepath.Base(file)
			idx := strings.LastIndex(name, ".")
			if idx < 0 {
				idx = len(name)
			}
			name = name[:idx]
			opts.envFiles = append(opts.envFiles, filepath.Join(filepath.Dir(file), name+".env"))
		}
	}
}

// WithEnv enables environment variable loading with the given prefix.
//
// Additional replacer pairs map config keys to env var names. For example,
// `WithEnv("APP", ".", "_")` maps `server.address` to `APP_SERVER_ADDRESS`.
func WithEnv(prefix string, replacerOldNewPairs ...string) Option {
	if len(replacerOldNewPairs)%2 != 0 {
		panic("replacer must in old-new pairs")
	}
	return func(opts *options) {
		opts.env = true
		opts.envPrefix = prefix
		for i := 0; i < len(replacerOldNewPairs); i += 2 {
			opts.envReplacers = append(opts.envReplacers, strings.NewReplacer(replacerOldNewPairs[i], replacerOldNewPairs[i+1]))
		}
	}
}

// WithEnvFile enables environment variable loading and preloads variables from
// the specified env file.
//
// This is useful when an application wants env semantics but still keeps local
// defaults in a checked-in or generated env file.
func WithEnvFile(file string, prefix string, replacerOldNewPairs ...string) Option {
	return func(opts *options) {
		WithEnv(prefix, replacerOldNewPairs...)(opts)
		opts.envFiles = append(opts.envFiles, file)
	}
}

// WithRemote configures Reader to load configuration from a remote provider.
//
// The provider arguments are passed through to Viper's remote configuration
// support. The caller is responsible for ensuring the required remote provider
// dependencies are present.
func WithRemote(typ string, provider, endpoint, path string) Option {
	return func(opts *options) {
		opts.remote = true
		opts.remoteType = typ
		opts.remoteProvider = provider
		opts.remoteEndpoint = endpoint
		opts.remotePath = path
	}
}

// WithRemoteSecured configures Reader to load configuration from a secured
// remote provider using the supplied secret.
func WithRemoteSecured(typ string, provider, endpoint, path string, secret string) Option {
	return func(opts *options) {
		WithRemote(typ, provider, endpoint, path)(opts)
		opts.remoteSecured = true
		opts.remoteSecuredSecret = secret
	}
}

// WithLogger sets the logger used for config discovery and env file loading.
func WithLogger(log Logger) Option {
	return func(opts *options) {
		opts.logger = log
	}
}

// WithOnChange registers a callback for config change events emitted by Viper.
func WithOnChange(f func(in fsnotify.Event)) Option {
	return func(opts *options) {
		opts.onChange = f
	}
}
