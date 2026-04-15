package util

import (
	"time"

	. "gbenson.net/go/zmachine/core"
)

// NewTicker returns a new [time.Ticker] with the given frequency.
func NewTicker(f Frequency) *time.Ticker {
	return time.NewTicker(time.Duration(f.Period() * float64(time.Second)))
}
