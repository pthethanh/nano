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

	Option func(*options)
)

// WithPaths sets config name, type, and search paths. Also loads env from <name>.env in each path.
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

// WithFile loads config from the given file and env from <fileName>.env in the same directory.
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

// WithEnv enables env loading with the given prefix and key replacers.
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

// WithEnvFile loads env from the specified file with prefix and key replacers.
func WithEnvFile(file string, prefix string, replacerOldNewPairs ...string) Option {
	return func(opts *options) {
		WithEnv(prefix, replacerOldNewPairs...)(opts)
		opts.envFiles = append(opts.envFiles, file)
	}
}

// WithRemote loads config from a remote provider.
func WithRemote(typ string, provider, endpoint, path string) Option {
	return func(opts *options) {
		opts.remoteType = typ
		opts.remoteProvider = provider
		opts.remoteEndpoint = endpoint
		opts.remotePath = path
	}
}

// WithRemoteSecured loads config from a secured remote provider.
func WithRemoteSecured(typ string, provider, endpoint, path string, secret string) Option {
	return func(opts *options) {
		WithRemote(typ, provider, endpoint, path)(opts)
		opts.remoteSecured = true
		opts.remoteSecuredSecret = secret
	}
}

// WithLogger sets a custom logger for config operations.
func WithLogger(log Logger) Option {
	return func(opts *options) {
		opts.logger = log
	}
}

// WithOnChange sets a callback for config change events.
func WithOnChange(f func(in fsnotify.Event)) Option {
	return func(opts *options) {
		opts.onChange = f
	}
}
