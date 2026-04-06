package domain

import "errors"

var ErrNoAvailablePort = errors.New("no available port in range")

// PortStrategy describes how to allocate a host port.
type PortStrategy struct {
	// Exactly one of these is set.
	Any   bool
	Exact uint32
	Range *PortRange
}

type PortRange struct {
	Min uint32
	Max uint32
}
