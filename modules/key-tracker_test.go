package modules

import (
	"testing"

	. "gbenson.net/go/zmachine/core"
	"gitlab.com/gomidi/midi/v2"
	"gotest.tools/v3/assert"
)

func TestKeyTracker(t *testing.T) {
	kt := &KeyTracker{}
	StartForTest(t, kt)

	// nothing received, not stepped
	assert.Equal(t, kt.Pitch(), 0*Hz)
	assert.Equal(t, kt.Velocity(), Fraction(0))
	assert.Equal(t, kt.Gate(), false)

	// note on received, not stepped
	kt.Receive(midi.NoteOn(0, 69-12, 123))

	assert.Equal(t, kt.Pitch(), 0*Hz)
	assert.Equal(t, kt.Velocity(), Fraction(0))
	assert.Equal(t, kt.Gate(), false)

	// received and stepped
	kt.Step()

	assert.Equal(t, kt.Pitch(), 220*Hz)
	assert.Equal(t, kt.Velocity(), Fraction(123.0/127))
	assert.Equal(t, kt.Gate(), true)

	// add a lower note (no change, higher note has priority)
	kt.Receive(midi.NoteOn(0, 48, 11))
	kt.Step()

	assert.Equal(t, kt.Pitch(), 220*Hz)
	assert.Equal(t, kt.Velocity(), Fraction(123.0/127))
	assert.Equal(t, kt.Gate(), true)

	// add a higher note (now we hears it)
	kt.Receive(midi.NoteOn(0, 60, 94))
	kt.Step()

	assert.Equal(t, int(kt.Pitch().Hz()), 261)
	assert.Equal(t, kt.Velocity(), Fraction(94.0/127))
	assert.Equal(t, kt.Gate(), true)

	// release the first note (no change, higher note still plays)
	kt.Receive(midi.NoteOff(0, 69-12))
	kt.Step()

	assert.Equal(t, int(kt.Pitch().Hz()), 261)
	assert.Equal(t, kt.Velocity(), Fraction(94.0/127))
	assert.Equal(t, kt.Gate(), true)

	// release the playing note (finally we hear the low note)
	kt.Receive(midi.NoteOff(0, 60))
	kt.Step()

	assert.Equal(t, int(kt.Pitch().Hz()), 130)
	assert.Equal(t, kt.Velocity(), Fraction(11.0/127))
	assert.Equal(t, kt.Gate(), true)

	// releasing non-playing notes has no effect
	for _, note := range []uint8{1, 2, 69 - 12, 60, 47, 49, 127} {
		kt.Receive(midi.NoteOff(0, note))
		kt.Step()

		assert.Equal(t, int(kt.Pitch().Hz()), 130)
		assert.Equal(t, kt.Velocity(), Fraction(11.0/127))
		assert.Equal(t, kt.Gate(), true)
	}

	// release the final note (with a velocity-0 note on)
	kt.Receive(midi.NoteOn(0, 48, 0))
	kt.Step()

	assert.Equal(t, int(kt.Pitch().Hz()), 130)         // floating
	assert.Equal(t, kt.Velocity(), Fraction(11.0/127)) // floating
	assert.Equal(t, kt.Gate(), false)
}
