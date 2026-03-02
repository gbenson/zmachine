package machine

// A Frequency is a float64 number of cycles per second.
type Frequency float64

// Frequently used units of frequency.
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
