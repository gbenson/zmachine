package core

// A Sample is a float64 number expected in the interval [-1.0,1.0].
type Sample float64

// Float64 returns s as a float64 without enforcing the interval.
func (s Sample) Float64() float64 {
	return float64(s)
}

// Clamped returns s clamped to the range [-1.0,1.0].
func (s Sample) Clamped() float64 {
	return max(-1, min(1, float64(s)))
}
