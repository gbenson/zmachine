package surface

import (
	"context"
	"time"

	"gbenson.net/go/logger"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
)

type EncoderID int
type PotID int

const (
	EncoderA EncoderID = iota
	EncoderB
	EncoderC
	EncoderD
	MenuEncoder
	EncoderX
	EncoderY
	EncoderZ
	NumEncoders
)

const (
	VolumePot PotID = iota
	NumPots
)

type Surface struct {
	log      *logger.Logger
	encoders [NumEncoders]encoder
	pots     [NumPots]pairedCC
	scanBuf  State
}

// Start implements [Starter].
func (s *Surface) Start(ctx context.Context) error {
	s.log = util.Logger(ctx, s)

	for i := range s.encoders {
		if EncoderID(i) == MenuEncoder {
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

func (s *Surface) Stop(ctx context.Context) {
	for i := range s.pots {
		defer s.pots[i].Stop(ctx)
	}
}

// Receive implements [zmachine.MIDISink].
func (s *Surface) Receive(msg midi.Message) {
	if s.receive(msg) {
		return
	}

	s.log.Warn().
		Hex("_msg", []byte(msg)).
		Stringer("msg", msg).
		Msg("Unhandled")
}

func (s *Surface) receive(msg midi.Message) bool {
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
		s.pots[VolumePot].receiveMSB(v)

	case cc == midi.VolumeLSB:
		s.pots[VolumePot].receiveLSB(v)

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

func (s *Surface) onEncoderMoved(n, amount int) {
	if n >= 0 && n < len(s.encoders) {
		s.encoders[n].receiveMovement(amount)
		return
	}

	s.log.Trace().
		Int("encoder", n).
		Int("moved", amount).
		Msg("Unhandled")
}

func (s *Surface) onEncoderClicked(n int, clicked bool) {
	if n >= 0 && n < len(s.encoders) {
		s.encoders[n].receiveEdge(clicked)
		return
	}

	s.log.Trace().
		Int("encoder", n).
		Bool("clicked", clicked).
		Msg("Unhandled")
}

type State struct {
	Encoders [NumEncoders]EncoderState
	Pots     [NumPots]PotState
}

type EncoderState struct {
	Delta float64 // Movement since last scan.
	Edges Edge    // Switch edges since last scan.
}

type PotState struct {
	Value float64 // Value when scanned.
	Delta float64 // Movement since last scan.
}

func (s *Surface) Scan() *State {
	cs := &s.scanBuf
	for i := range s.encoders { // do not copy!
		src := &s.encoders[i]
		dst := &cs.Encoders[i]

		dst.Delta = src.collectMovement()
		dst.Edges = src.collectEdges()
	}
	for i := range s.pots { // do not copy!
		dst := &cs.Pots[i]

		v := float64(s.pots[i].Load()) / ((1 << 14) - 1)
		dst.Delta = v - dst.Value
		dst.Value = v
	}
	return cs
}
