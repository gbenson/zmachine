package midi

import (
	. "gbenson.net/go/zmachine/core"
	"gitlab.com/gomidi/midi/v2"
)

type Router struct {
	DefaultReceiver  MIDISink
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
	if receiver := r.receiverFor(src, msg); receiver != nil {
		receiveMessage(receiver, src, msg)
	}
}

func (r *Router) receiverFor(deviceName string, msg midi.Message) MIDISink {
	var channel uint8
	switch {
	case deviceName == ControlSurfaceName:
		break

	case msg.GetChannel(&channel):
		return r.ChannelReceivers[channel]
	}

	return r.DefaultReceiver
}

func receiveMessage(r MIDISink, src string, msg midi.Message) {
	switch rr := r.(type) {
	case SourcedMIDISink:
		rr.ReceiveFrom(src, msg)
	default:
		r.Receive(msg)
	}
}
