// Package zmachine is a real-time audio synthesizer.
package zmachine

import (
	"context"
	"io"

	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/machine"
)

// A Machine manages a collection of components.
type Machine = machine.Machine

// New creates and initializes a new [Machine].
func New() *Machine {
	return machine.New()
}

// Open returns a new [io.Reader] reading from src.
func Open(ctx context.Context, src AudioSource) io.ReadCloser {
	return machine.NewReader(ctx, src)
}

// FromContext returns the [Machine] associated with ctx.
// It panics if ctx has no associated machine.
func FromContext(ctx context.Context) *Machine {
	return machine.FromContext(ctx)
}
