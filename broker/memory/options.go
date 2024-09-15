package memory

// Worker is an option to override the default number of worker and buffer.
func Worker[T any](worker, buffer int) Option[T] {
	return func(b *Broker[T]) {
		b.worker = worker
		b.buf = buffer
	}
}
