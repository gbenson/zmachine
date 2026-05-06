package core

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestZeroParameter(t *testing.T) {
	p := &Parameter{}

	assert.Equal(t, p.Load(), 0.0)

	for _, v := range []float64{-5, 22} {
		p.Store(v)
		assert.Equal(t, p.Load(), 0.0)
	}
}

func TestParameterClamping(t *testing.T) {
	p := &Parameter{Min: 1, Max: 4}

	p.Store(-5)
	assert.Equal(t, p.Load(), 1.0)

	p.Store(22)
	assert.Equal(t, p.Load(), 4.0)
}
