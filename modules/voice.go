package modules

// DON'T IMPLEMENT HERE:
//  - pitch bend (it should be *downstream*)
//  - glide (portamento; again, add it downstream)
//  - trigger (it's just a positive edge-detect on gate,
//    and can be added downstream if needed.
//  - ranges (keyboard splits) - it should be upstream (router)
//  - learn mode (you press a key and it uses the next note to see
//    what device/channel/etc) this voice is - again, add upstream
//    (router)
//  - polyphony - should be upstream, probably part of the router
//    but maybe a separate PolyVoice module.  Whatever it was would
//    likely manage a fixed number of MonoVoice objects rather than
//    reimplementing what's here (though maybe MonoVoice could be
//    split if tracking all notes is op for PolyVoice components)
//    using some kind of stealing algorithm for when it runs out of
//    voices like MIDI thing v2 does.
//
// DO IMPLEMENT HERE:
//  - Retrigger (bool) If true, make the gate go down for a short time
//      when a second Note ON is processed before the Note OFF of the
//      previous note arrives. By default the gate will only go down
//      after all Note OFFs have been processed.
//  - Other note selection modes (lowest, highest, most recent, least
//      recent, **arpeggiate**, etc)

import (
	"context"
	"math"
	"sync"
	"sync/atomic"

	"gbenson.net/go/logger"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
)

// Voice converts MIDI NoteOn and NoteOff messages into control
// signals for a monophonic voice.
type Voice struct {
	log *logger.Logger

	// Inner (internal, not-externally-visible) state:
	//  - updated by Receive
	//  - asynchronous (updates independently of Step)
	//  - likely to be updated by goroutines other than the stepper
	//  - multiple goroutines may update concurrently
	//  - all values private within Voice (i.e.: no accessors!)
	mu             sync.Mutex
	noteVelocities [128]uintptr

	// Outer (externally-visible) state:
	//  - updated by Step
	//  - synchronous (only updates when Stepped)
	//  - not concurrently updated (only one Step method of any module
	//    may be in-flight at any given time (other than in cases of
	//    nested modules where the nested module's Step is called by
	//    the Step of the module the nested module is nested in.)
	//  - some values are exposed (outside of Voice, via accessors).
	pitch       Frequency
	velocity    Fraction
	gate        bool
	lastOutputs uintptr

	// State transfer (inner to outer).
	// Updated by Receive, consumed by Step.
	moduleOutput atomic.Uintptr
}

// Start implements [util.Starter].
func (v *Voice) Start(ctx context.Context) error {
	v.log = util.Logger(ctx, v)
	return nil
}

// Receive implements [zmachine.MIDISink].
// It is safe for concurrent use.
func (v *Voice) Receive(msg midi.Message) {
	v.log.Trace().
		Hex("_msg", []byte(msg)).
		Stringer("msg", msg).
		Msg("Received")

	// Decode message.
	var note, vel uint8
	switch {
	case msg.GetNoteStart(nil, &note, &vel):
	case msg.GetNoteEnd(nil, &note):
	default:
		v.log.Warn().
			Hex("_msg", []byte(msg)).
			Stringer("msg", msg).
			Msg("Unhandled")
		return
	}

	if note > 127 {
		v.log.Error().
			Hex("_msg", []byte(msg)).
			Stringer("msg", msg).
			Str("shouldnt_be_possible", "gomidi v2.3.23 masks the bits").
			Str("so", "what changed?").
			Msg("Invalid note")
		return
	}

	// Update internal state.
	v.mu.Lock()
	v.noteVelocities[note] = uintptr(vel)

	// Interpret internal state.
	var outn, outv uintptr
	gate := false
	for n, v := range v.noteVelocities {
		if v < 1 {
			continue
		}
		outn = uintptr(n)
		outv = uintptr(v)
		gate = true
	}
	v.mu.Unlock()

	// Transfer internal state to module outputs.
	var outs uintptr
	if gate {
		outs = (outv << 8) | outn
	}
	v.moduleOutput.Store(outs)
}

func (v *Voice) Step() {
	outs := v.moduleOutput.Load()
	if outs == v.lastOutputs {
		return // unchanged
	}
	defer func() { v.lastOutputs = outs }()

	if outs == 0 {
		v.gate = false
		return // leave note and velocity floating
	}

	v.gate = true

	note := int(outs & 127)
	v.pitch = Frequency(440 * math.Pow(2, float64(note-69)/12))

	v.velocity = Fraction(float64(outs>>8) / 127)
}

// Pitch returns the frequency of the last played note.
func (v *Voice) Pitch() Frequency {
	return v.pitch
}

// Velocity returns the velocity of the last played note.
func (v *Voice) Velocity() Fraction {
	return v.velocity
}

// Gate returns true if a note is playing, false otherwise.
func (v *Voice) Gate() bool {
	return v.gate
}
