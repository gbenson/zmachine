package modules

import (
	"math"
	"testing"

	"gbenson.net/go/zmachine"
	"gotest.tools/v3/assert"
)

func TestPhaseAccumulator(t *testing.T) {
	pa := PhaseAccumulator{}
	assert.Equal(t, pa.timestep, 0.0)
	assert.Equal(t, pa.incr, 0.0)
	assert.Equal(t, pa.Phase(), 0.0)

	pa.timestep = zmachine.SampleRate(25000).Period() // XXX test Start()
	assert.Equal(t, pa.timestep, 40e-6)
	assert.Equal(t, pa.incr, 0.0)
	assert.Equal(t, pa.Phase(), 0.0)

	pa.SetFrequency(440)
	assert.Equal(t, pa.incr, 0.0176)
	assert.Equal(t, pa.Phase(), 0.0) // unchanged

	pa.Step()
	pa.Step()
	assert.Equal(t, pa.Phase(), 0.0352)

	pa.SetFrequency(100)
	assert.Equal(t, pa.incr, 0.004)
	assert.Equal(t, pa.Phase(), 0.0352) // unchanged

	pa.Step()
	assert.Equal(t, pa.Phase(), 0.0392)

	pa.Reset()
	assert.Equal(t, pa.Phase(), 0.0)
	assert.Equal(t, pa.incr, 0.004) // unchanged

	pa.Step()
	assert.Equal(t, pa.Phase(), 0.004)
	assert.Equal(t, pa.incr, 0.004) // unchanged

	assert.Equal(t, int(math.Round(pa.Phase()*1000)), 4)
	for _ = range 248 {
		pa.Step()
	}
	assert.Equal(t, int(math.Round(pa.Phase()*1000)), 996)

	pa.Step()
	assert.Equal(t, int(math.Round(pa.Phase()*1000)), 0)
	pa.Step()
	assert.Equal(t, int(math.Round(pa.Phase()*1000)), 4)
}

func TestFrequencyClipping(t *testing.T) {
	pa := &PhaseAccumulator{}
	pa.timestep = zmachine.SampleRate(44100).Period() // XXX test Start()

	assert.Equal(t, pa.incr, 0.0)

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
		pa.SetFrequency(float64(tc.set))
		assert.Check(t, 0 <= pa.incr)
		assert.Check(t, pa.incr <= 0.5)
		got := pa.Frequency()
		t.Logf("set: %d Hz, read: %.2f Hz", tc.set, got)
		assert.Equal(t, int(got), tc.want)
	}
}
