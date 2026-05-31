package surface

import (
	"math"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestInitialEncoderState(t *testing.T) {
	e := &encoder{}
	assert.Equal(t, e.lastV, 0.0)
	assert.Equal(t, e.lastT, time.Time{})
	assert.Equal(t, e.collectMovement(), 0.0)
	assert.Equal(t, e.collectEdges(), Edge(0))
}

// collecting zero movement doesn't update lastT or lastV
func TestCollectZeroMovement(t *testing.T) {
	e := &encoder{}

	assert.Equal(t, e.collectMovement(), 0.0)
	assert.Equal(t, e.lastV, 0.0)
	assert.Equal(t, e.lastT, time.Time{})

	testV := math.Float64frombits(0xdeadbeef98765432)
	testT := time.Unix(0x4e2fff94, 0x104bf2be)

	e.lastV = testV
	e.lastT = testT
	assert.Equal(t, e.collectMovement(), 0.0)
	assert.Equal(t, e.lastV, testV)
	assert.Equal(t, e.lastT, testT)
}
