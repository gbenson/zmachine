package sid

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestModelNumber(t *testing.T) {
	assert.Equal(t, Model6581.Number(), 6581)
	assert.Equal(t, Model8580.Number(), 8580)
}
