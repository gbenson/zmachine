package modules

import (
	"math"

	. "gbenson.net/go/zmachine/core"
)

// FractionShaper returns a [Shaper] that calls the given function.
func FractionShaper(f func(Fraction) Fraction) Shaper {
	return fractionShaper(f)
}

// fractionShaper is an adaptor that allows the use of ordinary
// functions as [Shaper]s.  If f is a function with the appropriate
// signature, then fractionShaper(f) is a [Shaper] that calls f.
type fractionShaper func(Fraction) Fraction

// Fraction implements [Shaper].
func (f fractionShaper) Fraction(x Fraction) Fraction {
	return f(x)
}

// Sample implements [Shaper].
func (f fractionShaper) Sample(x Fraction) Sample {
	return Sample(f(x)*2 - 1)
}

// SampleShaper returns a [Shaper] that calls the given function.
func SampleShaper(f func(Fraction) Sample) Shaper {
	return sampleShaper(f)
}

// sampleShaper is an adaptor that allows the use of ordinary
// functions as [Shaper]s.  If f is a function with the appropriate
// signature, then sampleShaper(f) is a [Shaper] that calls f.
type sampleShaper func(Fraction) Sample

// Fraction implements [Shaper].
func (f sampleShaper) Fraction(x Fraction) Fraction {
	return Fraction((f(x) + 1) / 2)
}

// Sample implements [Shaper].
func (f sampleShaper) Sample(x Fraction) Sample {
	return f(x)
}

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
