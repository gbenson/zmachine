package core

import "math"

// A Fraction is a float64 number expected in the unit interval [0.0,1.0].
type Fraction float64

// Float64 returns f as a float64 without enforcing the unit interval.
func (f Fraction) Float64() float64 {
	return float64(f)
}

// Clamped returns f clamped to the range [0.0,1.0].
func (f Fraction) Clamped() float64 {
	return max(0, min(1, float64(f)))
}

// Wrapped returns f reduced to the range [0.0,1.0) by wrapping modulo 1.
func (f Fraction) Wrapped() float64 {
	return float64(f) - math.Floor(float64(f))
}
