package logging

type (
	Option func(*options)

	options struct {
		logMethod   bool
		logRequest  bool
		logResponse bool
		logDuration bool
	}
)

func Method(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logMethod = enable
	}
}

func Request(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logRequest = enable
	}
}

func Response(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logResponse = enable
	}
}

func Duration(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logDuration = enable
	}
}

func All() Option {
	return func(o *options) {
		o.logRequest = true
		o.logResponse = true
		o.logDuration = true
		o.logMethod = true
	}
}

func newOpts(opts ...Option) *options {
	o := &options{
		logRequest:  false,
		logResponse: false,
		logDuration: false,
		logMethod:   false,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
