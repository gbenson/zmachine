package modules

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestEnvelopeDefaults(t *testing.T) {
	ctx := TestContext(t)
	e := Envelope{}
	assert.NilError(t, e.Start(ctx))

	assert.Equal(t, e.Attack.Duration.Load(), 0.0)
	assert.Equal(t, e.Decay.Duration.Load(), 0.0)
	assert.Equal(t, e.Sustain.Level.Load(), 1.0)
	assert.Equal(t, e.Decay.Duration.Load(), 0.0)
}
