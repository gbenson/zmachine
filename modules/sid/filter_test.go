package sid

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

func Test6581FilterCutoff(t *testing.T) {
	f := &Filter{Model: Model6581}
	StartForTest(t, f)

	f.SetFrequency(30 * Hz)
	assert.Equal(t, f.FC(), uint(0))
	assert.Equal(t, f.Frequency(), 30*Hz)

	f.SetFrequency(12000 * Hz)
	assert.Equal(t, f.FC(), uint(2047))
	assert.Equal(t, f.Frequency(), 12000*Hz)

	f.SetFC(0)
	assert.Equal(t, f.FC(), uint(0))
	assert.Equal(t, f.Frequency(), 30*Hz)

	f.SetFC(2047)
	assert.Equal(t, f.FC(), uint(2047))
	assert.Equal(t, f.Frequency(), 12000*Hz)
}

func Test8580FilterCutoff(t *testing.T) {
	f := &Filter{Model: Model8580}
	StartForTest(t, f)

	f.SetFrequency(0 * Hz)
	assert.Equal(t, f.FC(), uint(0))
	assert.Equal(t, f.Frequency(), 0*Hz)

	f.SetFrequency(12500 * Hz)
	assert.Equal(t, f.FC(), uint(2047))
	assert.Equal(t, f.Frequency(), 12500*Hz)

	f.SetFC(0)
	assert.Equal(t, f.FC(), uint(0))
	assert.Equal(t, f.Frequency(), 0*Hz)

	f.SetFC(2047)
	assert.Equal(t, f.FC(), uint(2047))
	assert.Equal(t, f.Frequency(), 12500*Hz)
}

func Test6581FilterDamping(t *testing.T) {
	f := &Filter{Model: Model6581}
	StartForTest(t, f)

	for resReg := range 16 {
		// this is the vibe-coded version I based mine on
		var expect float64
		inv := (^resReg) & 0x0F // Bitwise NOT, keep only 4 bits
		if inv > 0 {
			expect = float64(inv) / 8.0
		} else {
			expect = 0.06 // res=15 → near-self-oscillation
		}

		t.Logf("expect RES %2d => damping %.3f", resReg, expect)
		f.SetRES(uint(resReg))
		assert.Equal(t, f.damping, expect)
	}
}

func Test8580FilterDamping(t *testing.T) {
	f := &Filter{Model: Model8580}
	StartForTest(t, f)

	for _, tc := range []struct {
		resReg int
		expect float64
	}{
		{0, math.Sqrt(2)},
		{4, 1},
		{12, 0.5},
	} {
		resReg := tc.resReg
		expect := tc.expect

		t.Logf("expect RES %2d => damping %.3f", resReg, expect)
		f.SetRES(uint(resReg))
		assert.Equal(t, f.damping, expect)
	}
}
