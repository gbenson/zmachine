package core

import (
	"regexp"
	"strconv"
	"strings"
)

// A Frequency is a float64 number of cycles per second.
type Frequency float64

// Common used units of frequency.
const (
	Hz  Frequency = 1
	KHz           = 1000 * Hz
	BPM           = Hz / 60
)

// Hz returns the frequency as a floating point number of hertz.
func (f Frequency) Hz() float64 {
	return float64(f)
}

// BPM returns the frequency as a floating point number of beats/minute.
func (f Frequency) BPM() float64 {
	return float64(f) * 60
}

// Period returns the interval between cycles in seconds.
func (f Frequency) Period() float64 {
	if f == 0 {
		panic("division by zero")
	}
	return 1 / float64(f)
}

var frequencyRx = regexp.MustCompile(`^(.*?)(?:\s*([A-Za-z]+))?$`)
var freqUnitsByName = map[string]Frequency{
	"":    Hz,
	"hz":  Hz,
	"khz": KHz,
	"bpm": BPM,
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (f *Frequency) UnmarshalText(b []byte) error {
	s := string(b)
	m := frequencyRx.FindStringSubmatch(s)
	if len(m) != 3 {
		return InvalidFrequencyError(s)
	}

	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return InvalidFrequencyError(s)
	}

	u, ok := freqUnitsByName[strings.ToLower(m[2])]
	if !ok {
		return InvalidFrequencyError(s)
	}

	*f = Frequency(v) * u
	return nil
}
