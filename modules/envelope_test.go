package modules

import (
	"math"
	"testing"

	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/modules/envelope"
	"gotest.tools/v3/assert"
)

func TestEnvelopePhaseSequence(t *testing.T) {
	ctx := TestContext(t)
	e := Envelope[Sample]{}
	assert.NilError(t, e.Start(ctx))

	a := &e.phases[attackPhase]
	d := &e.phases[decayPhase]
	s := &e.phases[sustainPhase]
	r := &e.phases[releasePhase]
	o := &e.phases[noteOffPhase]

	assert.Equal(t, a.prev, o)
	assert.Equal(t, a.next, d)
	assert.Equal(t, d.prev, a)
	assert.Equal(t, d.next, s)
	assert.Equal(t, s.prev, d)
	assert.Equal(t, s.next, r)
	assert.Equal(t, r.prev, s)
	assert.Equal(t, r.next, o)
	assert.Equal(t, o.prev, r)
	assert.Equal(t, o.next, a)
}

func TestEnvelopeDefaults(t *testing.T) {
	ctx := TestContext(t)
	e := Envelope[Sample]{}
	assert.NilError(t, e.Start(ctx))

	assert.Equal(t, e.Attack().Duration(), 0.0)
	assert.Equal(t, e.Decay().Duration(), 0.0)
	assert.Equal(t, e.Sustain().Duration(), envelope.Hold)
	assert.Equal(t, e.Sustain().Level(), Sample(1))
	assert.Equal(t, e.Release().Duration(), 0.0)
}

func TestEnvelopePhaseLevelDeltas(t *testing.T) {
	ctx := TestContext(t)
	e := Envelope[Sample]{}
	assert.NilError(t, e.Start(ctx))

	assert.Equal(t, e.Sustain().Level(), Sample(1))

	assert.Equal(t, e.phases[attackPhase].levelDelta, Sample(1))
	assert.Equal(t, e.phases[decayPhase].levelDelta, Sample(0))
	assert.Equal(t, e.phases[sustainPhase].levelDelta, Sample(0))
	assert.Equal(t, e.phases[releasePhase].levelDelta, Sample(-1))
	assert.Equal(t, e.phases[noteOffPhase].levelDelta, Sample(0))

	e.Sustain().SetLevel(0.75)
	assert.Equal(t, e.Sustain().Level(), Sample(0.75))

	assert.Equal(t, e.phases[attackPhase].levelDelta, Sample(1))
	assert.Equal(t, e.phases[decayPhase].levelDelta, Sample(-0.25))
	assert.Equal(t, e.phases[sustainPhase].levelDelta, Sample(0))
	assert.Equal(t, e.phases[releasePhase].levelDelta, Sample(-0.75))
	assert.Equal(t, e.phases[noteOffPhase].levelDelta, Sample(0))

	e.Sustain().SetLevel(-0.5)
	assert.Equal(t, e.Sustain().Level(), Sample(-0.5))

	assert.Equal(t, e.phases[attackPhase].levelDelta, Sample(1))
	assert.Equal(t, e.phases[decayPhase].levelDelta, Sample(-1.5))
	assert.Equal(t, e.phases[sustainPhase].levelDelta, Sample(0))
	assert.Equal(t, e.phases[releasePhase].levelDelta, Sample(0.5))
	assert.Equal(t, e.phases[noteOffPhase].levelDelta, Sample(0))
}

func TestEnvelopePhaseDuration(t *testing.T) {
	ctx := TestContext(t)
	zmachine.FromContext(ctx).Config.Audio.SampleRate = 40 * KHz
	e := Envelope[Fraction]{}
	p := &e.phases[attackPhase]

	assert.Equal(t, p.timestep, 0.0)
	assert.Equal(t, p.incr, Fraction(0))
	assert.Equal(t, p.elapsed, Fraction(0))
	assert.Equal(t, p.Duration(), envelope.Hold)

	assert.NilError(t, e.Start(ctx))
	assert.Equal(t, p.timestep, 25e-6)
	assert.Check(t, p.incr >= 2)
	assert.Equal(t, p.elapsed, Fraction(0))
	assert.Equal(t, p.Duration(), 0.0)

	p.SetDuration(0.25)
	assert.Equal(t, p.timestep, 25e-6)
	assert.Equal(t, p.incr, Fraction(0.0001))
	assert.Equal(t, p.elapsed, Fraction(0))
	assert.Equal(t, p.Duration(), 0.25)

	p.SetDuration(-1e-9)
	assert.Equal(t, p.timestep, 25e-6)
	assert.Check(t, p.incr >= 2)
	assert.Equal(t, p.elapsed, Fraction(0))
	assert.Equal(t, p.Duration(), 0.0)
}

func TestEnvelopeOutputs(t *testing.T) {
	ctx := TestContext(t)
	e := Envelope[Sample]{}
	assert.NilError(t, e.Start(ctx))

	a := &e.phases[attackPhase]
	d := &e.phases[decayPhase]
	s := &e.phases[sustainPhase]
	r := &e.phases[releasePhase]
	o := &e.phases[noteOffPhase]

	a.SetDuration(1)
	d.SetDuration(1)
	r.SetDuration(1)

	for _, tc := range []struct {
		sustainLevel    Sample
		wantMidDecayOut Sample
	}{
		{1, 1},
		{0, 0.5},
		{-1, 0},
		{-0.5, 0.25},
		{0.5, 0.75},
	} {
		sustainLevel := tc.sustainLevel
		wantMidDecayOut := tc.wantMidDecayOut

		s.SetLevel(sustainLevel)
		assert.Equal(t, s.Level(), sustainLevel)

		for _, tc := range []struct {
			startPhase *envelopePhase[Sample]
			elapsed    Fraction
			wantOutput Sample
			wantPhase  *envelopePhase[Sample]
		}{
			{a, 0, 0, a},
			{a, 0.5, 0.5, a},
			{a, 1, 1, d},

			{d, 0, 1, d},
			{d, 0.5, wantMidDecayOut, d},
			{d, 1, sustainLevel, s},

			{s, 0, sustainLevel, s},
			{s, 0.5, sustainLevel, s},
			{s, 1, sustainLevel, s}, // incr = 0, i.e. it won't change

			{r, 0, sustainLevel, r},
			{r, 0.5, sustainLevel / 2, r},
			{r, 1, 0, o},

			{o, 0, 0, o},
			{o, 0.5, 0, o},
			{o, 1, 0, o}, // not a!
		} {
			p := tc.startPhase
			e.phase = p
			p.elapsed = tc.elapsed - Fraction(p.timestep)
			e.Step()
			assert.Equal(t, e.phase, tc.wantPhase)
			assert.Equal(t, e.Output(), tc.wantOutput)
		}
	}
}

func TestAmpEnvStepping(t *testing.T) {
	ctx := TestContext(t)
	zmachine.FromContext(ctx).Config.Audio.SampleRate = 1 * KHz
	const timestep = 1e-3

	for _, tc := range []struct {
		A           float64
		D           float64
		S           Fraction
		R           float64
		wantOutput  Fraction
		wantPhase   envelopePhaseIndex
		wantElapsed Fraction
	}{{
		// default settings
		A:          0,
		D:          0,
		S:          1,
		R:          0,
		wantOutput: 1,
		wantPhase:  noteOffPhase,
	}, {
		// output=S on degate should go straight to note off
		A:          0,
		D:          0,
		S:          0.8,
		R:          0,
		wantOutput: 0.8,
		wantPhase:  noteOffPhase,
	}, {
		// output=S on degate should go straight to note off even if R>0
		A:           0,
		D:           0,
		S:           0.8,
		R:           1,
		wantOutput:  0.8,
		wantPhase:   releasePhase,
		wantElapsed: 0,
	}, {
		A:           4 * timestep,
		D:           0,
		S:           1,
		R:           1,
		wantOutput:  0.25, // A/4
		wantPhase:   releasePhase,
		wantElapsed: 0.75,
	}, {
		A:           4 * timestep,
		D:           0,
		S:           0.8,
		R:           1,
		wantOutput:  0.25, // A/4
		wantPhase:   releasePhase,
		wantElapsed: (0.8 - 0.25) / 0.8,
	}} {
		e := Envelope[Fraction]{}

		assert.Check(t, e.phase == nil)
		assert.Equal(t, e.Gate(), false)
		assert.Equal(t, e.Output(), Fraction(0))

		assert.NilError(t, e.Start(ctx))
		assert.Equal(t, e.phase.index, noteOffPhase)
		assert.Equal(t, e.Gate(), false)
		assert.Equal(t, e.Output(), Fraction(0))

		e.Attack().SetDuration(tc.A)
		e.Decay().SetDuration(tc.D)
		e.Sustain().SetLevel(tc.S)
		e.Release().SetDuration(tc.R)

		assert.Equal(t, e.phase.index, noteOffPhase)
		assert.Equal(t, e.Gate(), false)
		assert.Equal(t, e.Output(), Fraction(0))

		// step (no note)
		e.Step()
		assert.Equal(t, e.phase.index, noteOffPhase)
		assert.Equal(t, e.Gate(), false)
		assert.Equal(t, e.Output(), Fraction(0))

		// note on
		e.SetGate(true)
		e.Step()
		assert.Check(t, e.phase.index != noteOffPhase)
		assert.Equal(t, e.Gate(), true)
		assert.Equal(t, e.Output(), tc.wantOutput)

		// note off
		e.SetGate(false)
		e.Step()
		assert.Equal(t, e.phase.index, tc.wantPhase)
		assert.Equal(t, e.Gate(), false)
		if tc.wantPhase == noteOffPhase {
			assert.Equal(t, e.Output(), Fraction(0))
		} else {
			assert.Equal(t, tc.wantPhase, releasePhase)
			p := e.phase
			wantElapsed := tc.wantElapsed + p.incr
			haveElapsed := p.elapsed
			delta := wantElapsed - haveElapsed
			assert.Check(t, math.Abs(delta.Float64()) < 1e-12)
		}
	}
}
