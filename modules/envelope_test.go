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

func TestEnvelopeParameters(t *testing.T) {
	ctx := TestContext(t)
	e := Envelope{}
	params := e.Parameters()

	assert.NilError(t, e.Start(ctx))
	attack := params.Get("attack")
	assert.Equal(t, attack.Load(), 0.0)

	attack.Store(0.75)
	assert.Equal(t, attack.Load(), 0.75)

	assert.Equal(t, e.Attack.Duration.Load(), 0.75)
	assert.Equal(t, e.Decay.Duration.Load(), 0.0)
	assert.Equal(t, e.Sustain.Level.Load(), 1.0)
	assert.Equal(t, e.Decay.Duration.Load(), 0.0)
}
