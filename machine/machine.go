package machine

import (
	"context"
	"time"

	"gbenson.net/go/logger"
)

const DefaultSampleRate Frequency = 48 * KHz
const DefaultMaxLatency = 10 * time.Millisecond

// A Machine manages a collection of components.
type Machine struct {
	SampleRate Frequency
	MaxLatency time.Duration
}

// New creates and initializes a new [Machine].
func New() *Machine {
	return &Machine{
		SampleRate: DefaultSampleRate,
		MaxLatency: DefaultMaxLatency,
	}
}

// contextKey is a value for use with context.WithValue. It's used
// as a pointer so it fits in an interface{} without allocation.
type contextKey struct{}

// machineKey is a context key that can be used in zmachine-managed
// code to access the managing [Machine]. The associated value will
// be of type *Machine.
var machineKey = &contextKey{}

// WithContext returns a copy of ctx with the receiver attached.
func (m *Machine) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, machineKey, m)
}

// FromContext returns the [Machine] associated with ctx.
// It panics if ctx has no associated machine.
func FromContext(ctx context.Context) *Machine {
	machine, _ := ctx.Value(machineKey).(*Machine)
	if machine == nil {
		panic("nil machine")
	}
	return machine
}

// TestContext returns its receiver's context after associating a
// [logger.Logger] and a semi-configured [Machine] with it.  It's
// intended for use with [testing.T].
func TestContext(t logger.Contexter) context.Context {
	return New().WithContext(logger.TestContext(t))
}
