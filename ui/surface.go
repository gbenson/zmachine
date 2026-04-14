package ui

import (
	"context"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
)

type Surface struct {
	log *logger.Logger
}

func (s *Surface) Start(ctx context.Context) error {
	s.log = util.Logger(ctx, s)
	return nil
}

// Receive implements [zmachine.MIDISink].
func (s *Surface) Receive(msg midi.Message) {
	s.log.Trace().
		Hex("_msg", []byte(msg)).
		Stringer("msg", msg).
		Msg("Received")
}
