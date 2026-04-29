package envelope

import (
	"math"

	. "gbenson.net/go/zmachine/core"
)

type Type interface {
	Fraction | Sample
}

type Slope interface {
	Shape(x Fraction) Fraction

	// Unshape is the inverse function of shape.
	Unshape(y Fraction) Fraction
}

type Phase interface {
	Duration() float64
	SetDuration(d float64)
}

var Hold = math.Inf(+1)

type SlopePhase interface {
	Phase
	Slope() Slope
	SetSlope(s Slope)
}

type LevelPhase[T Type] interface {
	Phase
	Level() T
	SetLevel(level T)
}

type linearSlope struct{}

func (s *linearSlope) Shape(x Fraction) Fraction   { return x }
func (s *linearSlope) Unshape(y Fraction) Fraction { return y }

var LinearSlope Slope = &linearSlope{}
