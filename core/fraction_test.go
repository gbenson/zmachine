package core

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

func TestFractionClamping(t *testing.T) {
	for _, tc := range []struct {
		input, expectClamped float64
	}{
		{-1e6, 0},
		{-1e-6, 0},
		{-0, 0},
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
		f := Fraction(tc.input)
		assert.Equal(t, f.Float64(), tc.input)
		assert.Equal(t, f.Clamped(), tc.expectClamped)
	}
}

func TestFractionWrapping(t *testing.T) {
	for _, tc := range []struct {
		input, expectWrapped float64
	}{
		{0, 0},
		{1e-6, 1e-6},
		{0.3, 0.3},
		{math.Ln2, math.Ln2},
		{1 - 1e-8, 0.99999999},
		{1, 0},
		{1 + 1e-8, 1e-8},
		{1.001, 0.001},
		{math.Pi, math.Pi - 3},
		{5.2, 0.2},
		{123465.78, 0.78},
		{-0.1, 0.9},
		{-0.8, 0.2},
		{-1.1, 0.9},
		{-20.4, 0.6},
	} {
		f := Fraction(tc.input)
		assert.Equal(t, f.Float64(), tc.input)
		t.Log("want:", tc.expectWrapped, "got:", f.Wrapped())
		assert.Check(t, math.Abs(f.Wrapped()-tc.expectWrapped) < 1e-11)
	}
}
