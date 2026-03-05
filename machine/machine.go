package machine

import (
	"context"
	"errors"
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
	Source     AudioSource
	Sink       AudioSink

	started bool

	ctx  context.Context
	stop context.CancelFunc

	metrics *readerMetrics
}

// contextKey is a value for use with context.WithValue. It's used
// as a pointer so it fits in an interface{} without allocation.
type contextKey struct{}

// machineKey is a context key that can be used in zmachine-managed
// code to access the managing [Machine]. The associated value will
// be of type *Machine.
var machineKey = &contextKey{}

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
	} else if m.ctx != nil {
		return errors.New("already started")
	}

	m.init(ctx)
	ctx = m.ctx

	if src, ok := m.Source.(util.Starter); ok {
		if err := util.LoggedStart(ctx, src); err != nil {
			return err
		}
	}

	r := &reader{ctx: ctx, source: m.Source}
	m.metrics = &r.metrics

	if err := m.Sink.Start(ctx, r); err != nil {
		defer m.deferableLoggedClose(m.Source)
		return err
	}

	m.started = true
	return nil
}

func (m *Machine) init(ctx context.Context) {
	if m.SampleRate <= 0 {
		m.SampleRate = DefaultSampleRate
	}
	if m.MaxLatency <= 0 {
		m.MaxLatency = DefaultMaxLatency
	}

	ctx = context.WithValue(ctx, machineKey, m)
	ctx, m.stop = context.WithCancel(ctx)
	m.ctx = ctx
}

// TestContext returns its receiver's context after associating a
// [logger.Logger] and a semi-configured [Machine] with it.  It's
// intended for use with [testing.T].
func TestContext(t logger.Contexter) context.Context {
	ctx := logger.TestContext(t)
	m := &Machine{}
	m.init(ctx)
	return m.ctx
}

// Close implements [io.Closer].
func (m *Machine) Close() error {
	if !m.started {
		return errors.New("never started")
	} else if m.ctx == nil {
		panic("nil context")
	}

	defer func() {
		if rm := m.metrics; rm != nil {
			rm.logReport(util.Logger(m.ctx, rm))
		}
	}()

	if stop := m.stop; stop != nil {
		stop()
	}

	var closers []io.Closer
	for _, comp := range []any{m.Source, m.Sink} {
		if c, ok := comp.(io.Closer); ok {
			closers = append(closers, c)
		}
	}

	return util.ForEach(closers, m.loggedClose)
}

func (m *Machine) loggedClose(c io.Closer) error {
	return util.LoggedClose(m.ctx, c)
}

func (m *Machine) deferableLoggedClose(comp any) {
	if c, ok := comp.(io.Closer); ok {
		util.DeferableLoggedClose(m.ctx, c)
	}
}
