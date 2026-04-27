package midi

import (
	. "gbenson.net/go/zmachine/core"
	"gitlab.com/gomidi/midi/v2"
)

type Router struct {
	GlobalReceiver   MIDISink
	ChannelReceivers [16]MIDISink
}

// A SourcedMIDISink is a component that receives MIDI messages
// from named devices.
type SourcedMIDISink interface {
	MIDISink

	// ReceiveFrom receives one MIDI message from a named device.
	ReceiveFrom(src string, msg midi.Message)
}

// Receive implements [MIDISink].
func (r *Router) Receive(msg midi.Message) {
	r.ReceiveFrom("", msg)
}

// ReceiveFrom implements [SourcedMIDISink].
func (r *Router) ReceiveFrom(src string, msg midi.Message) {
	receiveMessage(r.GlobalReceiver, src, false, msg)

	var channel uint8
	if !msg.GetChannel(&channel) {
		return
	}

	if r := r.ChannelReceivers[channel]; r != nil {
		receiveMessage(r, src, false, msg)
	}
}

// ReceiveFromSurface implements [ControlSink].
func (r *Router) ReceiveFromSurface(msg midi.Message) {
	if cs, ok := r.GlobalReceiver.(ControlSink); ok {
		cs.ReceiveFromSurface(msg)
	}
}

func receiveMessage(
	r MIDISink,
	src string,
	srcIsControlSurface bool,
	msg midi.Message,
) {
	if srcIsControlSurface {
		if rr, ok := r.(ControlSink); ok {
			rr.ReceiveFromSurface(msg)
			return
		}
	}

	if src != "" {
		if rr, ok := r.(SourcedMIDISink); ok {
			rr.ReceiveFrom(src, msg)
			return
		}
	}

	r.Receive(msg)
}
