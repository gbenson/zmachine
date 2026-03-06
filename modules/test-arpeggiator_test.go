package modules

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestTestArpeggiatorTempo(t *testing.T) {
	ctx := testContext(t)

	ta := &TestArpeggiator{Receiver: &TestMIDISink{ctx}}
	assert.NilError(t, ta.Start(ctx))

	tempo := ta.Tempo()
	assert.Equal(t, tempo.Period(), 10.0/7)
	assert.Equal(t, tempo.Hz(), 0.7)
}
