package logging

// Option is a function that configures logging interceptor options.
type Option func(*options)

type (
	options struct {
		logMethod   bool
		logRequest  bool
		logResponse bool
		logDuration bool
	}
)

// Method returns an Option that enables or disables logging of gRPC method names.
// When called without arguments or with true, it enables method logging.
// When called with false, it disables method logging.
func Method(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logMethod = enable
	}
}

// Request returns an Option that enables or disables logging of gRPC request data.
// When called without arguments or with true, it enables request logging.
// When called with false, it disables request logging.
func Request(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logRequest = enable
	}
}

// Response returns an Option that enables or disables logging of gRPC response data.
// When called without arguments or with true, it enables response logging.
// When called with false, it disables response logging.
func Response(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logResponse = enable
	}
}

// Duration returns an Option that enables or disables logging of gRPC call duration.
// When called without arguments or with true, it enables duration logging.
// When called with false, it disables duration logging.
func Duration(enabled ...bool) Option {
	enable := len(enabled) == 0 || len(enabled) > 0 && enabled[0]
	return func(o *options) {
		o.logDuration = enable
	}
}

// All returns an Option that enables all logging options:
// method, request, response, and duration.
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
