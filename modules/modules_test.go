package modules

import (
	"context"

	"gbenson.net/go/zmachine/machine"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
)

// testContext returns its receiver's context after associating
// a [logger.Logger] and a semi-configured [Machine] with it.
var testContext = machine.TestContext

type TestMIDISink struct {
	ctx context.Context
}

func (ts *TestMIDISink) Receive(msg midi.Message) {
	util.Logger(ts.ctx, ts).Debug().
		Stringer("msg", msg).
		Msg("Received")
}
