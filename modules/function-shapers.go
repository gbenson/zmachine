package modules

import (
	"math"

	. "gbenson.net/go/zmachine/core"
)

// SineShaper generates a sine wave.
var SineShaper = SampleShaper(func(x Fraction) Sample {
	return Sample(math.Sin(x.Float64() * 2 * math.Pi))
})

// TriangleShaper generates a triangle wave with the same phase as
// [Sine] (positive in the first half-cycle, negative in the second.)
var TriangleShaper = SampleShaper(func(x Fraction) Sample {
	y := Sample(x.Wrapped() * 4)
	switch int(y) {
	case 0:
		return y
	default: // 1, 2
		return 2 - y
	case 3:
		return y - 4
	}
})

// FallingSawShaper generates a sawtooth waveform with the same sign
// as [Sine] (positive in the first half-cycle, negative in the
// second.)
var FallingSawShaper = FractionShaper(func(x Fraction) Fraction {
	return Fraction(1 - x.Float64())
})

// RisingSawShaper generates a sawtooth waveform with the opposite sign as
// [Sine] (negative in the first half-cycle, positive in the second.)
var RisingSawShaper = FractionShaper(func(x Fraction) Fraction {
	return x
})
