package zmachine

import (
	"context"
	"io"

	"gitlab.com/gomidi/midi/v2"
)

// An AudioSink is a component that consumes samples.
type AudioSink interface {
	// Start causes the sink to consume samples until cancelled.
	Start(ctx context.Context, r io.Reader) error
}

// A MIDISink is a component that receives MIDI messages.
type MIDISink interface {
	// Receive receives one MIDI message.
	Receive(midi.Message)
}
