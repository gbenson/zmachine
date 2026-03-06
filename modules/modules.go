package modules

import "gbenson.net/go/zmachine"

// A Frequency is a float64 number of cycles per second.
type Frequency = zmachine.Frequency

// Frequently used units of frequency.
const (
	Hz  = zmachine.Hz
	KHz = zmachine.KHz
	BPM = zmachine.BPM
)

// A MIDISink is a component that receives MIDI messages.
type MIDISink = zmachine.MIDISink
