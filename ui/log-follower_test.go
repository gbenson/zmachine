package ui

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestShortString(t *testing.T) {
	for _, tc := range []struct {
		l, m, c, s, want string
	}{
		{"info", "Blown", "ui.Trumpet", "", "ui.Trumpet: blown"},
		{"info", "Seen", "midi.Follower", "fancy kbd", "fancy kbd: seen"},
		{"info", "Seen", "midi.Follower", "Midi Through", ""},
		{"info", "x", "midi.Follower", "Control Surface", "ui.Surface: x"},
	} {
		assert.Equal(t, shortString(tc.l, tc.m, tc.c, tc.s), tc.want)
	}
}
