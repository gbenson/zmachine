package modules

import (
	"context"
	"math"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
)

// Voice converts MIDI NoteOn and NoteOff messages into control
// signals for a monophonic voice.
type Voice struct {
	log  *logger.Logger
	note uint8
}

// Start implements [util.Starter].
func (v *Voice) Start(ctx context.Context) error {
	v.log = util.Logger(ctx, v)
	return nil
}

// Receive implements [zmachine.MIDISink].
func (v *Voice) Receive(msg midi.Message) {
	v.log.Trace().
		Hex("_msg", []byte(msg)).
		Stringer("msg", msg).
		Msg("Received")

	_ = msg.GetNoteStart(nil, &v.note, nil)
}

func (v *Voice) Pitch() Frequency {
	return Frequency(440 * math.Pow(2, float64(int(v.note)-69)/12))
}
