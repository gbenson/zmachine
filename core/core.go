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

// A ControlSink is a component that receives MIDI messages from
// Zmachine control surfaces.
type ControlSink interface {
	MIDISink

	// ReceiveFrom receives one MIDI message from a named device.
	ReceiveFromSurface(msg midi.Message)
}

// A Clamper is a value with an acceptable range it can clamp itself to.
type Clamper interface {
	// Clamped returns the clamped value.
	Clamped() float64
}

// A Shaper is a component that samples a periodic waveform.
type Shaper interface {
	// Fraction returns the amplitude of its waveform at the given
	// position in its cycle as a [Fraction].
	Fraction(x Fraction) Fraction

	// Sample returns the amplitude of its waveform at the given
	// position in its cycle as a [Sample].
	Sample(x Fraction) Sample
}

// A Starter is a component that must be started before use.
type Starter interface {
	// Start causes a component to run until cancelled.
	Start(ctx context.Context) error
}
