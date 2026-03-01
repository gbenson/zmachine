package machine

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSampleRatePeriod(t *testing.T) {
	r := SampleRate(12500)
	assert.Equal(t, r.Period(), 80e-6)
}

func TestNilSampleRate(t *testing.T) {
	var recovered any
	defer func() {
		assert.Equal(t, recovered, "division by zero")
	}()
	defer func() {
		recovered = recover()
	}()

	var r SampleRate
	_ = r.Period()
}
