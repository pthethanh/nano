package log

func WithContextRetrievers(ctxs ...ContextRetriever) Option {
	return func(l *Logger) {
		l.ctxs = ctxs
	}
}
