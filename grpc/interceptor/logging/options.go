package logging

type (
	Option func(*options)

	options struct {
		logMethod   bool
		logRequest  bool
		logReply    bool
		logDuration bool

		appendToContext AppendToContextFunc
	}
)

func LogMethod(f AppendToContextFunc) Option {
	return func(o *options) {
		o.logMethod = true
		o.appendToContext = f
	}
}

func LogRequest(enabled bool) Option {
	return func(o *options) {
		o.logRequest = enabled
	}
}

func LogReply(enabled bool) Option {
	return func(o *options) {
		o.logReply = enabled
	}
}

func LogDuration(enabled bool) Option {
	return func(o *options) {
		o.logDuration = enabled
	}
}

func newOpts(opts ...Option) *options {
	o := &options{
		logRequest:  true,
		logReply:    true,
		logDuration: true,

		logMethod:       true,
		appendToContext: nil,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
