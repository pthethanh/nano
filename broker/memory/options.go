package memory

// Worker is an option to override the default number of worker and buffer.
//
// worker is clamped to a minimum of 1: a broker with zero worker goroutines
// has nothing to drain the internal channel, so every Publish call would
// block forever once the buffer fills. buffer is clamped to a minimum of 0.
func Worker[T any](worker, buffer int) Option[T] {
	return func(b *Broker[T]) {
		if worker < 1 {
			worker = 1
		}
		if buffer < 0 {
			buffer = 0
		}
		b.worker = worker
		b.buf = buffer
	}
}
