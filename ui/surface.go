package ui

import (
	"context"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
)

type encoderID int

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

type surface struct {
	log      *logger.Logger
	encoders [numEncoders]encoder
	scanbuf  collectedState
}

func (s *surface) init(ctx context.Context) {
	s.log = util.Logger(ctx, s)
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
		s.onVolumeMSB(int(v))

	case cc == midi.VolumeLSB:
		s.onVolumeLSB(int(v))

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

func (s *surface) onVolumeMSB(v int) {
	s.log.Trace().
		Int("volume_msb", v).
		Msg("Unhandled")
}

func (s *surface) onVolumeLSB(v int) {
	s.log.Trace().
		Int("volume_lsb", v).
		Msg("Unhandled")
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
	encoderDeltas [numEncoders]int
	encoderEdges  [numEncoders]Edge
	volumeValue   int
	volumeDelta   int
}

func (s *surface) Scan() *collectedState {
	cs := &s.scanbuf
	for i := range s.encoders { // do not copy!
		cs.encoderDeltas[i] = s.encoders[i].collectMovement()
		cs.encoderEdges[i] = s.encoders[i].collectEdges()
	}
	// XXX volume!
	return cs
}
