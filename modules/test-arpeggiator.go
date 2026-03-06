package modules

import (
	"context"

	"gitlab.com/gomidi/midi/v2"
)

// TestArpeggiator emits the notes of the Stranger Things arpeggio.
type TestArpeggiator struct {
	Receiver MIDISink
	Channel  uint8
	Velocity uint8

	pa           PhaseAccumulator
	beatsPerNote Frequency
	notes        []uint8
	lastNote     uint8
}

// Start implements [zmachine.Starter].
func (ta *TestArpeggiator) Start(ctx context.Context) error {
	if ta.Receiver == nil {
		panic("nil receiver")
	}

	if err := ta.pa.Start(ctx); err != nil {
		return err
	}

	ta.notes = []uint8{48, 52, 55, 59, 60, 59, 55, 52}

	const tempo = 84 * BPM // this from score: ":quarter-note:=84"
	ta.beatsPerNote = 0.5  // this scales tempo to step on eigth notes.
	ta.SetTempo(tempo)

	return nil
}

func (ta *TestArpeggiator) Tempo() Frequency {
	return ta.pa.Frequency()
}

func (ta *TestArpeggiator) SetTempo(t Frequency) {
	ta.pa.SetFrequency(t * ta.beatsPerNote)
}

func (ta *TestArpeggiator) Step() {
	ta.pa.Step()

	notes := ta.notes
	note := notes[int(ta.pa.Phase()*float64(len(ta.notes)))]

	lastNote := ta.lastNote
	if note == lastNote {
		return
	}
	defer func() { ta.lastNote = note }()

	emit := ta.Receiver.Receive
	channel := ta.Channel

	velocity := ta.Velocity
	if velocity == 0 {
		velocity = 127
	}

	if lastNote != 0 {
		emit(midi.NoteOff(channel, lastNote))
	}
	emit(midi.NoteOn(channel, note, velocity))
}
