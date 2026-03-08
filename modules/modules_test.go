package modules

import (
	"context"

	"gbenson.net/go/zmachine/util"
	"gbenson.net/go/zmachine/util/testutil"
	"gitlab.com/gomidi/midi/v2"
)

var (
	StartForTest = testutil.StartForTest
	TestContext  = testutil.TestContext
)

type TestMIDISink struct {
	ctx context.Context
}

func (ts *TestMIDISink) Receive(msg midi.Message) {
	util.Logger(ts.ctx, ts).Debug().
		Stringer("msg", msg).
		Msg("Received")
}
