package machine

// A SampleRate is an integer number of samples per second.
type SampleRate uint32

// Period returns the interval between samples in seconds.
func (r SampleRate) Period() float64 {
	if r == 0 {
		panic("division by zero")
	}
	return 1.0 / float64(r)
}
