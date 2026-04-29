package core

import "context"

// StarterFunc is an adaptor that allows the use of ordinary
// functions as starters.  If f is a function with the appropriate
// signature, then StarterFunc(f) is a [Starter] that calls f.
type StarterFunc func(context.Context) error

// Start implements [Starter].
func (f StarterFunc) Start(ctx context.Context) error {
	return f(ctx)
}

// FractionShaper is an adapter that allows the use of ordinary
// functions as shapers.  If f is a function with the appropriate
// signature, then FractionShaper(f) is a [Shaper] that calls f.
type FractionShaper func(Fraction) Fraction

// Fraction implements [Shaper].
func (f FractionShaper) Fraction(x Fraction) Fraction {
	return f(x)
}

// Sample implements [Shaper].
func (f FractionShaper) Sample(x Fraction) Sample {
	return Sample(f(x)*2 - 1)
}

// SampleShaper is an adapter that allows the use of ordinary
// functions as shapers.  If f is a function with the appropriate
// signature, then SampleShaper(f) is a [Shaper] that calls f.
type SampleShaper func(Fraction) Sample

// Fraction implements [Shaper].
func (f SampleShaper) Fraction(x Fraction) Fraction {
	return Fraction((f(x) + 1) / 2)
}

// Sample implements [Shaper].
func (f SampleShaper) Sample(x Fraction) Sample {
	return f(x)
}
