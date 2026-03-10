package core

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

func TestSampleClamping(t *testing.T) {
	for _, tc := range []struct {
		input, expectClamped float64
	}{
		{-1e6, -1},
		{-math.Sqrt2, -1},
		{-math.Sqrt2 / 2, -math.Sqrt2 / 2},
		{-1e-6, -1e-6},
		{-0, -0},
		{0, 0},
		{1e-6, 1e-6},
		{0.3, 0.3},
		{0.3, 0.3},
		{math.Ln2, math.Ln2},
		{1 - 1e-8, 0.99999999},
		{1, 1},
		{1 + 1e-8, 1},
		{1.001, 1},
		{math.Pi, 1},
		{5, 1},
		{1e20, 1},
	} {
		s := Sample(tc.input)
		assert.Equal(t, s.Float64(), tc.input)
		assert.Equal(t, s.Clamped(), tc.expectClamped)
	}
}
