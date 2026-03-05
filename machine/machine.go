package machine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/util"
)

const DefaultSampleRate Frequency = 48 * KHz
const DefaultMaxLatency = 10 * time.Millisecond

// A Machine manages a collection of components.
type Machine struct {
	SampleRate Frequency
	MaxLatency time.Duration
	Source     io.Reader
	Sink       AudioSink

	started bool

	ctx  context.Context
	stop context.CancelFunc
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

func (m *Machine) Start(ctx context.Context) error {
	if ctx == nil {
		panic("nil context")
	} else if m.Source == nil {
		panic("nil source")
	} else if m.Sink == nil {
		panic("nil sink")
	} else if m.SampleRate <= 0 {
		return fmt.Errorf("%v: invalid sample rate", m.SampleRate)
	} else if m.MaxLatency <= 0 {
		return fmt.Errorf("%v: invalid max latency", m.MaxLatency)
	} else if m.ctx != nil {
		return errors.New("already started")
	}

	ctx, m.stop = context.WithCancel(m.WithContext(ctx))
	m.ctx = ctx

	if err := m.Sink.Start(ctx, m.Source); err != nil {
		return err
	}

	m.started = true
	return nil
}

// TestContext returns its receiver's context after associating a
// [logger.Logger] and a semi-configured [Machine] with it.  It's
// intended for use with [testing.T].
func TestContext(t logger.Contexter) context.Context {
	return New().WithContext(logger.TestContext(t))
}

// Close implements [io.Closer].
func (m *Machine) Close() error {
	if !m.started {
		return errors.New("never started")
	} else if m.ctx == nil {
		panic("nil context")
	}

	if stop := m.stop; stop != nil {
		stop()
	}

	if c, ok := m.Sink.(io.Closer); ok {
		defer util.DeferableLoggedClose(m.ctx, c)
	}

	return nil
}
