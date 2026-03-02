package modules

import (
	"context"
	"math"
)

// TestArpeggiator emits the notes of the Stranger Things arpeggio.
type TestArpeggiator struct {
	pa    PhaseAccumulator
	freqs []Frequency
	scale Frequency
}

// Start implements [zmachine.Starter].
func (ta *TestArpeggiator) Start(ctx context.Context) error {
	if err := ta.pa.Start(ctx); err != nil {
		return err
	}

	for _, nn := range []int{48, 52, 55, 59, 60, 59, 55, 52} {
		freq := Frequency(440 * math.Pow(2, float64(nn-69)/12))
		ta.freqs = append(ta.freqs, freq)
	}
	ta.scale = 0.5 // they're eigth-notes

	ta.SetTempo(84 * BPM)

	return nil
}

func (ta *TestArpeggiator) Tempo() Frequency {
	return ta.pa.Frequency()
}

func (ta *TestArpeggiator) SetTempo(t Frequency) {
	ta.pa.SetFrequency(t * ta.scale)
}

func (ta *TestArpeggiator) Frequency() Frequency {
	freqs := ta.freqs
	return freqs[int(ta.pa.Phase()*float64(len(freqs)))]
}

func (ta *TestArpeggiator) Step() {
	ta.pa.Step()
}
