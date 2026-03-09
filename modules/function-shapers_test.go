package modules

import (
	"math"
	"testing"

	. "gbenson.net/go/zmachine/core"
	"gotest.tools/v3/assert"
)

func TestFallingSawShaperFraction(t *testing.T) {
	testWaveform := []int{
		8, 7, 6, 5, 4, 3, 2, 1,
	}
	for n, scaledY := range testWaveform {
		x := Fraction(float64(n) / float64(len(testWaveform)))
		y := FallingSawShaper.Fraction(x)
		t.Logf("n=%d, x=%6.4f, y=%4.2f", n, x.Float64(), y.Float64())
		assert.Equal(t, y.Float64(), float64(scaledY)/8)
	}
}

func TestRisingSawShaperSample(t *testing.T) {
	testWaveform := []int{
		-4, -3, -2, -1, 0, 1, 2, 3,
	}
	for n, scaledY := range testWaveform {
		x := Fraction(float64(n) / float64(len(testWaveform)))
		y := RisingSawShaper.Sample(x)
		t.Logf("n=%d, x=%6.4f, y=%4.2f", n, x.Float64(), y.Float64())
		assert.Equal(t, y.Float64(), float64(scaledY)/4)
	}
}

func TestSineShaperSample(t *testing.T) {
	r3o2 := math.Sqrt(3) / 2
	testWaveform := []float64{
		0, 0.5, r3o2,
		1, r3o2, 0.5,
		0, -0.5, -r3o2,
		-1, -r3o2, -0.5,
	}
	for n, expectY := range testWaveform {
		x := Fraction(float64(n) / float64(len(testWaveform)))
		y := SineShaper.Sample(x)
		t.Logf("n=%d, x=%6.4f, y=%4.2f", n, x.Float64(), y.Float64())
		delta := expectY - y.Float64()
		assert.Assert(t, math.Abs(delta) < 1e-12)
	}
}

func TestTriangleShaperFraction(t *testing.T) {
	testWaveform := []int{
		4, 5, 6, 7,
		8, 7, 6, 5,
		4, 3, 2, 1,
		0, 1, 2, 3,
	}
	for n, scaledY := range testWaveform {
		x := Fraction(float64(n) / float64(len(testWaveform)))
		y := TriangleShaper.Fraction(x)
		t.Logf("n=%d, x=%6.4f, y=%4.2f", n, x.Float64(), y.Float64())
		assert.Equal(t, y.Float64(), float64(scaledY)/8)
	}
}

func TestTriangleShaperSample(t *testing.T) {
	testWaveform := []int{
		0, 1, 2, 3,
		4, 3, 2, 1,
		0, -1, -2, -3,
		-4, -3, -2, -1,
	}
	for n, scaledY := range testWaveform {
		x := Fraction(float64(n) / float64(len(testWaveform)))
		y := TriangleShaper.Sample(x)
		t.Logf("n=%d, x=%6.4f, y=%4.2f", n, x.Float64(), y.Float64())
		assert.Equal(t, y.Float64(), float64(scaledY)/4)
	}
}
