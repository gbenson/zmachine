package ui

import (
	"fmt"
	"io"
	"strings"

	"gbenson.net/go/zmachine/ui/internal/logfollower"
)

// NewLogFollower creates and initializes a new log follower.
func NewLogFollower(w io.Writer) io.WriteCloser {
	return logfollower.New(w)
}

type logRecord struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Component string `json:"comp"`
	Source    string `json:"source"`
}

func (r *logRecord) ShortString() string {
	return shortString(r.Level, r.Message, r.Component, r.Source)
}

func shortString(l, m, c, s string) string {
	// fast path for normal operation
	if l == "info" && m != "" {
		switch c {
		case "":
			return m
		case "midi.Follower":
			switch s {
			case "Midi Through":
				return ""
			default:
				c = s
			}
		}
		return fmt.Sprintf("%s: %s", c, strings.ToLower(m))
	}

	// non-normal messages
	switch l {
	case "warn":
		l = "warning"
	case "fatal":
		l = "fatal error"
	}

	if c == "" {
		c = "unattributed"
	}

	return fmt.Sprintf("%s %s!", c, l)
}
