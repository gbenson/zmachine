package core

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestFrequencyPeriod(t *testing.T) {
	f := Frequency(12500)
	assert.Equal(t, f.Period(), 80e-6)
}

func TestFrequencyBPM(t *testing.T) {
	f := 135 * BPM
	assert.Equal(t, f.Hz(), 2.25)
	assert.Equal(t, f.Period(), 60.0/135)
}

func TestNilFrequency(t *testing.T) {
	var recovered any
	defer func() {
		assert.Equal(t, recovered, "division by zero")
	}()
	defer func() {
		recovered = recover()
	}()

	var f Frequency
	_ = f.Period()
}
