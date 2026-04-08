package ui

import (
	"fmt"
	"io"

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
}

func (r *logRecord) ShortString() string {
	return shortString(r.Level, r.Message, r.Component)
}

func shortString(l, m, c string) string {
	// fast path for normal operation
	if l == "info" && m != "" {
		if c == "" {
			return m
		}
		return fmt.Sprintf("%s %s", m, c)
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
