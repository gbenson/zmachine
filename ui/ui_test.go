package ui

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestInitialPage(t *testing.T) {
	var ui UI
	assert.Equal(t, ui.CurrentPage(), &ui.loggerPage)
}
