// Package zmachine is a real-time audio synthesizer.
package zmachine

import (
	"context"
	"errors"
	"io"

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

// New creates and initializes a new [Machine].
func New() *Machine {
	return machine.New()
}

// Open returns a new [io.Reader] reading from src.
func Open(ctx context.Context, src AudioSource) io.ReadCloser {
	return machine.NewReader(ctx, src)
}

// Run runs a [Machine] until interrupted or cancelled.
func Run(ctx context.Context, m *Machine) error {
	if err := m.Start(ctx); err != nil {
		return err
	}
	defer util.DeferableLoggedClose(ctx, m)

	<-ctx.Done()

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// FromContext returns the [Machine] associated with ctx.
// It panics if ctx has no associated machine.
func FromContext(ctx context.Context) *Machine {
	return machine.FromContext(ctx)
}
