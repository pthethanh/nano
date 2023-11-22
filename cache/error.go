package cache

import "errors"

var (
	// ErrNotFound is an error report that the key is not found.
	ErrNotFound = errors.New("cache: key not found")

	// ErrInValidConnState indicates that the cache has not Open/initialized accordingly yet.
	ErrInValidConnState = errors.New("cache: invalid connection state")
)
