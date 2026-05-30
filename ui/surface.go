package ui

import (
	"context"
	"time"

	"gbenson.net/go/logger"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
)

type encoderID int
type potID int

const (
	encoderA encoderID = iota
	encoderB
	encoderC
	encoderD
	menuEncoder
	freqEncoder
	resEncoder
	modEncoder
	numEncoders
)

const (
	volumePot potID = iota
	numPots
)

type surface struct {
	log      *logger.Logger
	encoders [numEncoders]encoder
	pots     [numPots]pairedCC
	scanbuf  collectedState
}

// Start implements [Starter].
func (s *surface) Start(ctx context.Context) error {
	s.log = util.Logger(ctx, s)

	for i := range s.encoders {
		if encoderID(i) == menuEncoder {
			continue
		}
		s.encoders[i].setAcceleration(500 * time.Millisecond)
	}

	pots := make([]Starter, len(s.pots))
	for i := range s.pots {
		pots[i] = &s.pots[i]
	}
	return util.StartAll(ctx, pots)
}

func (s *surface) Stop(ctx context.Context) {
	for i := range s.pots {
		defer s.pots[i].Stop(ctx)
	}
}

// Receive implements [zmachine.MIDISink].
func (s *surface) Receive(msg midi.Message) {
	if s.receive(msg) {
		return
	}

	s.log.Warn().
		Hex("_msg", []byte(msg)).
		Stringer("msg", msg).
		Msg("Unhandled")
}

func (s *surface) receive(msg midi.Message) bool {
	var cc, v uint8
	if !msg.GetControlChange(nil, &cc, &v) {
		return false
	}

	const encoderStart = midi.GeneralPurposeSlider1
	const encoderLimit = midi.BankSelectLSB
	const N = encoderLimit - encoderStart
	const encoderSwitchLimit = midi.AllSoundOff
	const encoderSwitchStart = encoderSwitchLimit - N

	switch {
	case cc == midi.VolumeMSB:
		s.pots[volumePot].receiveMSB(v)

	case cc == midi.VolumeLSB:
		s.pots[volumePot].receiveLSB(v)

	case cc >= encoderStart && cc < encoderLimit:
		n := int(cc - encoderStart)
		s.onEncoderMoved(n, int(v)-64)

	case cc >= encoderSwitchStart && cc < encoderSwitchLimit:
		n := int(cc - encoderSwitchStart)
		s.onEncoderClicked(n, v > 63)

	default:
		return false
	}

	return true
}

func (s *surface) onEncoderMoved(n, amount int) {
	if n >= 0 && n < len(s.encoders) {
		s.encoders[n].receiveMovement(amount)
		return
	}

	s.log.Trace().
		Int("encoder", n).
		Int("moved", amount).
		Msg("Unhandled")
}

func (s *surface) onEncoderClicked(n int, clicked bool) {
	if n >= 0 && n < len(s.encoders) {
		s.encoders[n].receiveEdge(clicked)
		return
	}

	s.log.Trace().
		Int("encoder", n).
		Bool("clicked", clicked).
		Msg("Unhandled")
}

type collectedState struct {
	encoderDeltas [numEncoders]float64
	encoderEdges  [numEncoders]Edge
	potValues     [numPots]float64
	potDeltas     [numPots]float64
}

func (s *surface) Scan() *collectedState {
	cs := &s.scanbuf
	for i := range s.encoders { // do not copy!
		cs.encoderDeltas[i] = s.encoders[i].collectMovement()
		cs.encoderEdges[i] = s.encoders[i].collectEdges()
	}
	for i := range s.pots { // likewise, don't copy
		v := float64(s.pots[i].Load()) / ((1 << 14) - 1)
		cs.potDeltas[i] = cs.potValues[i] - v
		cs.potValues[i] = v
	}
	return cs
}
