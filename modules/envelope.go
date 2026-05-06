package modules

import (
	"context"
	"math"

	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
)

type envelopePhase int

const (
	noteOffPhase envelopePhase = iota
	attackPhase
	decayPhase
	sustainPhase
	releasePhase
)

var envelopePhaseName = map[envelopePhase]string{
	attackPhase:  "attack",
	decayPhase:   "decay",
	sustainPhase: "sustain",
	releasePhase: "release",
	noteOffPhase: "note-off",
}

type Envelope struct {
	Attack  EnvelopeSlopePhase
	Decay   EnvelopeSlopePhase
	Sustain EnvelopeLevelPhase
	Release EnvelopeSlopePhase

	Rate           Parameter     // how fast do we advance through the envelope?
	phase          envelopePhase // what phase are we in?
	gate, lastGate bool
	level          float64
}

type EnvelopeSlopePhase struct {
	Duration Parameter
}

type EnvelopeLevelPhase struct {
	Level Parameter
}

// Start implements [Starter].
func (e *Envelope) Start(ctx context.Context) error {
	machine := zmachine.FromContext(ctx)

	const maxDuration = 30
	e.Attack.Duration.Max = maxDuration
	e.Decay.Duration.Max = maxDuration
	e.Release.Duration.Max = maxDuration

	e.Sustain.Level.Max = 1
	e.Sustain.Level.Store(1)

	rate := machine.Config.Audio.SampleRate.Hz()
	e.Rate.Min = rate / 100 // arbitrary
	e.Rate.Max = rate * 100 // likewise
	e.Rate.Store(rate)

	return nil
}

// Parameters implements [Parameterized].
func (e *Envelope) Parameters() Parameters {
	return Parameters{
		"attack":  &e.Attack.Duration,
		"decay":   &e.Decay.Duration,
		"sustain": &e.Sustain.Level,
		"release": &e.Release.Duration,
		"rate":    &e.Rate,
	}
}

func (e *Envelope) Gate() bool {
	return e.gate
}

func (e *Envelope) SetGate(v bool) {
	e.gate = v
}

func (e *Envelope) Level() float64 {
	return e.level
}

func (e *Envelope) Step() {
	if e.gate {
		if !e.lastGate {
			// Note on
			e.phase = attackPhase
			e.lastGate = e.gate
		}
	} else if e.lastGate {
		// Note off
		e.phase = releasePhase
		e.lastGate = e.gate
	}

	timestep := 1 / e.Rate.Load()

	switch e.phase {
	case attackPhase:
		d := e.Attack.Duration.Load()
		if d > timestep {
			incr := timestep / d
			e.level += incr
			if e.level < 1 {
				break
			}
		}
		e.level = 1
		e.phase = decayPhase

	case decayPhase:
		d := e.Decay.Duration.Load()
		target := e.Sustain.Level.Load()
		if d > timestep {
			incr := (target - 1) * timestep / d
			e.level += incr
			if e.level > target {
				break
			}
		}
		e.level = target
		e.phase = sustainPhase

	case releasePhase:
		d := e.Release.Duration.Load()
		if d > timestep {
			origin := e.Sustain.Level.Load()
			incr := origin * timestep / d
			if math.Signbit(incr) == math.Signbit(e.level) {
				incr *= -1
			}
			e.level += incr
			if math.Abs(e.level) > 1e-6 {
				break
			}
		}
		e.level = 0
		e.phase = noteOffPhase
	}
}
