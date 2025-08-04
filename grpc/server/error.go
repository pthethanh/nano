package server

import "errors"

// ErrUnknownServiceType is returned when the service type is not recognized.
var (
	ErrUnknownServiceType = errors.New("unknown service type")
)
