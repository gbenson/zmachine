package modules

import (
	"math"
	"testing"

	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
	"gotest.tools/v3/assert"
)

func TestPhaseAccumulator(t *testing.T) {
	ctx := TestContext(t)
	zmachine.FromContext(ctx).SampleRate = 25 * KHz

	pa := PhaseAccumulator{}
	assert.Equal(t, pa.timestep, 0.0)
	assert.Equal(t, pa.incr, Fraction(0))
	assert.Equal(t, pa.Phase(), Fraction(0))

	assert.NilError(t, pa.Start(ctx))
	assert.Equal(t, pa.timestep, 40e-6)
	assert.Equal(t, pa.incr, Fraction(0))
	assert.Equal(t, pa.Phase(), Fraction(0))

	pa.SetFrequency(440)
	assert.Equal(t, pa.incr, Fraction(0.0176))
	assert.Equal(t, pa.Phase(), Fraction(0)) // unchanged

	pa.Step()
	pa.Step()
	assert.Equal(t, pa.Phase(), Fraction(0.0352))

	pa.SetFrequency(100)
	assert.Equal(t, pa.incr, Fraction(0.004))
	assert.Equal(t, pa.Phase(), Fraction(0.0352)) // unchanged

	pa.Step()
	assert.Equal(t, pa.Phase(), Fraction(0.0392))

	pa.Reset()
	assert.Equal(t, pa.Phase(), Fraction(0))
	assert.Equal(t, pa.incr, Fraction(0.004)) // unchanged

	pa.Step()
	assert.Equal(t, pa.Phase(), Fraction(0.004))
	assert.Equal(t, pa.incr, Fraction(0.004)) // unchanged

	assert.Equal(t, int(math.Round(pa.Phase().Float64()*1000)), 4)
	for _ = range 248 {
		pa.Step()
	}
	assert.Equal(t, int(math.Round(pa.Phase().Float64()*1000)), 996)

	pa.Step()
	assert.Equal(t, int(math.Round(pa.Phase().Float64()*1000)), 0)
	pa.Step()
	assert.Equal(t, int(math.Round(pa.Phase().Float64()*1000)), 4)
}

func TestFrequencyClipping(t *testing.T) {
	ctx := TestContext(t)
	zmachine.FromContext(ctx).SampleRate = Frequency(44100)

	pa := &PhaseAccumulator{}
	assert.NilError(t, pa.Start(ctx))

	assert.Equal(t, pa.incr, Fraction(0))

	for _, tc := range []struct {
		set, want int
	}{
		{0, 0},
		{-14, 0},
		{16537, 16537},
		{22049, 22049},
		{22050, 22050},
		{22051, 22049},
		{22052, 22048},
		{44098, 2},
		{44099, 1},
		{44100, 0},
		{44101, 1},
		{44102, 2},
		{44103, 3},
	} {
		pa.SetFrequency(Frequency(tc.set))
		assert.Check(t, 0 <= pa.incr)
		assert.Check(t, pa.incr <= 0.5)
		got := pa.Frequency()
		t.Logf("set: %d Hz, read: %.2f Hz", tc.set, got)
		assert.Equal(t, int(got), tc.want)
	}
}
