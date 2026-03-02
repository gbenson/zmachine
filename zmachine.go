// Package zmachine is a real-time audio synthesizer.
package zmachine

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/machine"
	"gbenson.net/go/zmachine/util"
)

// A Machine manages a collection of components.
type Machine = machine.Machine

// An AudioSource is a component that generates samples.
type AudioSource = machine.AudioSource

// An AudioSink is a component that consumes samples.
type AudioSink = machine.AudioSink

// A Frequency is a float64 number of cycles per second.
type Frequency = machine.Frequency

// Frequently used units of frequency.
const (
	Hz  = machine.Hz
	KHz = machine.KHz
	BPM = machine.BPM
)

// ContextKey is a context key that can be used in zmachine-managed
// code to access the managing [Machine]. The associated value will
// be of type *Machine.
var ContextKey = machine.ContextKey

// New creates and initializes a new [Machine].
func New() *Machine {
	return &Machine{}
}

// Run runs a [Machine] until interrupted or cancelled.
func Run(ctx context.Context, m *Machine) error {
	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := m.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// Logger returns a [logger.Logger] suitable for comp.
func Logger(ctx context.Context, comp any) *logger.Logger {
	return util.ComponentLogger(ctx, comp)
}

// FromContext returns the [Machine] associated with ctx.
// It panics if ctx has no associated machine.
func FromContext(ctx context.Context) *Machine {
	machine, _ := ctx.Value(ContextKey).(*Machine)
	if machine == nil {
		panic("nil machine")
	}
	return machine
}
