package modules

import (
	"testing"

	"gitlab.com/gomidi/midi/v2"
	"gotest.tools/v3/assert"
)

func TestVoice(t *testing.T) {
	v := &Voice{}
	StartForTest(t, v)

	// nothing received, not stepped
	assert.Equal(t, v.Pitch(), 0*Hz)
	assert.Equal(t, v.Velocity(), 0.0)
	assert.Equal(t, v.Gate(), false)

	// note on received, not stepped
	v.Receive(midi.NoteOn(0, 69-12, 123))

	assert.Equal(t, v.Pitch(), 0*Hz)
	assert.Equal(t, v.Velocity(), 0.0)
	assert.Equal(t, v.Gate(), false)

	// received and stepped
	v.Step()

	assert.Equal(t, v.Pitch(), 220*Hz)
	assert.Equal(t, v.Velocity(), 123.0/127)
	assert.Equal(t, v.Gate(), true)

	// add a lower note (no change, higher note has priority)
	v.Receive(midi.NoteOn(0, 48, 11))
	v.Step()

	assert.Equal(t, v.Pitch(), 220*Hz)
	assert.Equal(t, v.Velocity(), 123.0/127)
	assert.Equal(t, v.Gate(), true)

	// add a higher note (now we hears it)
	v.Receive(midi.NoteOn(0, 60, 94))
	v.Step()

	assert.Equal(t, int(v.Pitch().Hz()), 261)
	assert.Equal(t, v.Velocity(), 94.0/127)
	assert.Equal(t, v.Gate(), true)

	// release the first note (no change, higher note still plays)
	v.Receive(midi.NoteOff(0, 69-12))
	v.Step()

	assert.Equal(t, int(v.Pitch().Hz()), 261)
	assert.Equal(t, v.Velocity(), 94.0/127)
	assert.Equal(t, v.Gate(), true)

	// release the playing note (finally we hear the low note)
	v.Receive(midi.NoteOff(0, 60))
	v.Step()

	assert.Equal(t, int(v.Pitch().Hz()), 130)
	assert.Equal(t, v.Velocity(), 11.0/127)
	assert.Equal(t, v.Gate(), true)

	// releasing non-playing notes has no effect
	for _, note := range []uint8{1, 2, 69 - 12, 60, 47, 49, 127} {
		v.Receive(midi.NoteOff(0, note))
		v.Step()

		assert.Equal(t, int(v.Pitch().Hz()), 130)
		assert.Equal(t, v.Velocity(), 11.0/127)
		assert.Equal(t, v.Gate(), true)
	}

	// release the final note (with a velocity-0 note on)
	v.Receive(midi.NoteOn(0, 48, 0))
	v.Step()

	assert.Equal(t, int(v.Pitch().Hz()), 130) // floating
	assert.Equal(t, v.Velocity(), 11.0/127)   // floating
	assert.Equal(t, v.Gate(), false)
}
