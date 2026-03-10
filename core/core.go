// Package core provides core functionality used throughout zmachine.
package core

import (
	"context"
	"io"

	"gitlab.com/gomidi/midi/v2"
)

// An AudioSource is a component that generates samples.
type AudioSource interface {
	// Generate stores up to len(buf) samples into buf.
	Generate(ctx context.Context, buf []float32) (n int, err error)
}

// An AudioSink is a component that consumes samples.
type AudioSink interface {
	// Start causes the sink to consume samples until cancelled.
	Start(ctx context.Context, r io.Reader) error
}

// A MIDISink is a component that receives MIDI messages.
type MIDISink interface {
	// Receive receives one MIDI message.
	Receive(msg midi.Message)
}
