package machine

import (
	"context"
	"io"
	"time"

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

	ctx  context.Context
	stop context.CancelFunc

	metrics *readerMetrics
}

// contextKey is a value for use with context.WithValue. It's used
// as a pointer so it fits in an interface{} without allocation.
type contextKey struct{}

// ContextKey is a context key that can be used in zmachine-managed
// code to access the managing [Machine]. The associated value will
// be of type *Machine.
var ContextKey = &contextKey{}

func (m *Machine) Start(ctx context.Context) error {
	if m.Source == nil {
		panic("nil source")
	} else if m.Sink == nil {
		panic("nil sink")
	}

	if m.SampleRate <= 0 {
		m.SampleRate = DefaultSampleRate
	}
	if m.MaxLatency <= 0 {
		m.MaxLatency = DefaultMaxLatency
	}

	ctx = context.WithValue(ctx, ContextKey, m)
	ctx, m.stop = context.WithCancel(ctx)
	m.ctx = ctx

	if src, ok := m.Source.(util.Starter); ok {
		if err := util.LoggedStart(ctx, src); err != nil {
			return err
		}
	}

	r := &reader{ctx: ctx, source: m.Source}
	m.metrics = &r.metrics

	return m.Sink.Start(ctx, r)
}

// Close implements [io.Closer].
func (m *Machine) Close() error {
	defer func() {
		if rm := m.metrics; rm != nil {
			rm.logReport(util.ComponentLogger(m.ctx, rm))
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
