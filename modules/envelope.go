package modules

import (
	"context"
	"math"

	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/modules/envelope"
)

type envelopePhaseIndex int

const (
	attackPhase envelopePhaseIndex = iota
	decayPhase
	sustainPhase
	releasePhase
	noteOffPhase // must be > releasePhase
	numEnvelopePhases
)

var envelopePhaseName = map[envelopePhaseIndex]string{
	attackPhase:  "attack",
	decayPhase:   "decay",
	sustainPhase: "sustain",
	releasePhase: "release",
	noteOffPhase: "note-off",
}

type Envelope[T envelope.Type] struct {
	phases         [numEnvelopePhases]envelopePhase[T]
	phase          *envelopePhase[T]
	gate, lastGate bool
	output         T
}

type envelopePhase[T envelope.Type] struct {
	prev, next *envelopePhase[T]  // sequence (doubly-linked loop)
	index      envelopePhaseIndex // position in sequence
	timestep   float64            // how much does time advance when we Step()?
	incr       Fraction           // how much do we add to elapsed when we Step()
	elapsed    Fraction           // accumulated time spent in this phase
	slope      envelope.Slope     // output shaper
	level      T                  // target level (i.e. level at phase exit)
	levelDelta T                  // change in level across phase
}

// Start implements [Starter].
func (e *Envelope[T]) Start(ctx context.Context) error {
	machine := zmachine.FromContext(ctx)
	ts := machine.Config.Audio.SampleRate.Period()

	n := len(e.phases)
	for i := range n {
		p := &e.phases[i]
		p.prev = &e.phases[(i+n-1)%n]
		p.next = &e.phases[(i+1)%n]
		p.index = envelopePhaseIndex(i)
		p.timestep = ts
		p.SetSlope(envelope.LinearSlope)
	}

	e.Attack().SetDuration(0)
	e.phases[attackPhase].SetLevel(1)

	e.Decay().SetDuration(0)

	e.Sustain().SetDuration(envelope.Hold)
	e.Sustain().SetLevel(1)

	e.Release().SetDuration(0)

	e.phase = &e.phases[noteOffPhase]

	return nil
}

func (e *Envelope[T]) Attack() envelope.SlopePhase {
	return &e.phases[attackPhase]
}

func (e *Envelope[T]) Decay() envelope.SlopePhase {
	return &e.phases[decayPhase]
}

func (e *Envelope[T]) Sustain() envelope.LevelPhase[T] {
	return &e.phases[sustainPhase]
}

func (e *Envelope[T]) Release() envelope.SlopePhase {
	return &e.phases[releasePhase]
}

// Durations < timestep/maxIncr are considered zero.
const maxIncr = 10

func (p *envelopePhase[T]) Duration() float64 {
	switch {
	case p.incr <= 0:
		return envelope.Hold
	case p.incr >= maxIncr:
		return 0
	default:
		return p.timestep / p.incr.Float64()
	}
}

func (p *envelopePhase[T]) SetDuration(d float64) {
	if d > math.MaxFloat64 {
		p.incr = 0 // hold
		return
	} else if d*maxIncr < p.timestep {
		// If both attack and decay are zero then notes jump straight
		// to sustain.  Exactly the same thing happens if attack and
		// decay sum to less than a timestep, with the advantage that
		// you don't need to care about or check for dividing by zero.
		p.incr = Fraction(maxIncr) // XXX report clipping?
	} else {
		p.incr = Fraction(p.timestep / d)
	}
}

func (p *envelopePhase[T]) Level() T {
	return p.level
}

type clampable interface {
	Clamped() float64
}

func (p *envelopePhase[T]) SetLevel(v T) {
	if vv, ok := any(v).(Clamper); ok {
		// XXX report clipping, if any?
		v = T(vv.Clamped())
	}
	p.level = v

	if p.index == sustainPhase {
		p.prev.SetLevel(v)
		p.next.SetLevel(0)
	}
	p.levelDelta = v - p.prev.level
}

func (p *envelopePhase[T]) Slope() envelope.Slope {
	return p.slope
}

func (p *envelopePhase[T]) SetSlope(s envelope.Slope) {
	p.slope = s
}

func (e *Envelope[T]) Gate() bool {
	return e.gate
}

func (e *Envelope[T]) SetGate(v bool) {
	e.gate = v
}

func (e *Envelope[T]) Output() T {
	return e.output
}

func (p *envelopePhase[T]) String() string {
	return envelopePhaseName[p.index]
}

func (e *Envelope[T]) Step() {
	p := e.phase

	if e.gate {
		if !e.lastGate {
			// Note on
			p = &e.phases[attackPhase]
			e.phase = p
			p.elapsed = 0
		}
	} else if e.lastGate {
		// Note off
		if p.index < releasePhase {
			p = &e.phases[releasePhase]
			e.phase = p
			if p.levelDelta == 0 {
				// happens if sustain level is 0
				p.elapsed = 1
			} else {
				p.elapsed = p.slope.Unshape(
					Fraction((e.output - p.prev.level) / p.levelDelta))
			}
		}
	}
	e.lastGate = e.gate

	todo := Fraction(1)
	for p.index != noteOffPhase {
		incr := p.incr

		elapsed := p.elapsed + incr*todo
		if elapsed < 1 {
			e.output = p.prev.level + T(p.slope.Shape(elapsed))*p.levelDelta
			p.elapsed = elapsed
			return
		}

		todo = (elapsed - 1) / incr
		e.output = p.level
		p = p.next
		e.phase = p
		p.elapsed = 0
	}
}
