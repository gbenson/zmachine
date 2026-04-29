package modules

import (
	"context"
	"math"

	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
)

type PhaseAccumulator struct {
	timestep float64  // param: how much does time advance when we Step()?
	incr     Fraction // input: how much do we add to phase when we Step()
	phase    Fraction // output: the accumulated phase; range: [0..1)
}

// Start implements [Starter].
func (pa *PhaseAccumulator) Start(ctx context.Context) error {
	machine := zmachine.FromContext(ctx)
	pa.timestep = machine.Config.Audio.SampleRate.Period()
	return nil
}

func (pa *PhaseAccumulator) Frequency() Frequency {
	return Frequency(pa.incr.Float64() / pa.timestep)
}

func (pa *PhaseAccumulator) SetFrequency(f Frequency) {
	hz := f.Hz()
	if hz < 0 {
		// XXX report clipping?
		pa.incr = 0
		return
	}

	incr := pa.timestep * hz
	if incr > 0.5 {
		// XXX report aliasing?
		incr -= math.Floor(incr)
		if incr > 0.5 {
			incr = 1 - incr
		}
	}

	pa.incr = Fraction(incr)
}

func (pa *PhaseAccumulator) Phase() Fraction {
	return pa.phase
}

func (pa *PhaseAccumulator) Step() {
	phase := pa.phase + pa.incr
	if phase >= 1 {
		phase -= 1
	}
	pa.phase = phase
}

// Reset resets the accumulated phase to zero.
func (pa *PhaseAccumulator) Reset() {
	pa.phase = 0
}
