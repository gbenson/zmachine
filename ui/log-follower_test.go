package ui

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestShortString(t *testing.T) {
	for _, tc := range []struct {
		l, m, c, want string
	}{
		{"info", "Blew", "ui.Trumpet", "Blew ui.Trumpet"},
	} {
		assert.Equal(t, shortString(tc.l, tc.m, tc.c), tc.want)
	}
}
