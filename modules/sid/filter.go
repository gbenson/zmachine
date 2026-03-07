package sid

import (
	"context"

	"gbenson.net/go/zmachine"
)

type Filter struct {
	Model Model

	sampleRate  Frequency
	model       model
	fcreg       uint    // 11-bit value (0..2047) derived from F0,F1
	resreg      float64 // 4-bit vaue (0..15) from RES, as float64
	cutoffTable [2048]float64
	w0          float64 // vibe: Filter coefficient (2*sin(π*freq/sampleRate))
	damping     float64 // vibe: Damping factor for resonance
	input       float64
	inputBias   float64 // vibe: DC offset (biases NMOS inverters)

	// vibe: subtle asymmetric saturation models NMOS inverter character
	outputCurve func(float64) float64

	lp float64 // Low-pass output
	hp float64 // High-pass output
	bp float64 // Band-pass output
}

// Start implements [zmachine.Starter].
func (f *Filter) Start(ctx context.Context) error {
	if f.Model == nil {
		panic("nil model")
	} else if model, ok := f.Model.(model); !ok {
		panic("invalid model")
	} else {
		f.model = model
	}

	machine := zmachine.FromContext(ctx)
	f.sampleRate = machine.SampleRate

	f.model.InitCutoffTable(f.cutoffTable[:], f.sampleRate.Hz())
	f.inputBias = f.model.FilterInputBias()
	f.outputCurve = f.model.FilterOutputCurve()

	return nil
}

func (f *Filter) Frequency() Frequency {
	return Frequency(f.model.FCtoHz(f.fcreg))
}

func (ff *Filter) SetFrequency(f Frequency) {
	ff.SetFC(ff.model.HzToFC(f.Hz()))
}

func (f *Filter) FC() uint {
	return f.fcreg
}

func (f *Filter) SetFC(fc uint) {
	fc &= 2047
	f.fcreg = fc
	f.w0 = f.cutoffTable[fc]
}

func (f *Filter) Resonance() float64 {
	return f.resreg / 15
}

func (f *Filter) SetResonance(r float64) {
	f.setRES(max(0, min(15, r*15)))
}

func (f *Filter) RES() uint {
	return uint(f.resreg)
}

func (f *Filter) SetRES(res uint) {
	f.setRES(float64(res & 15))
}

func (f *Filter) setRES(res float64) {
	f.resreg = res
	f.damping = f.model.DampingForRES(res)
}

func (f *Filter) Input() float64 {
	return f.input
}

func (f *Filter) SetInput(v float64) {
	f.input = v
}

func (f *Filter) Step() {
	input := f.input
	w0 := f.w0
	lp := f.lp
	hp := f.hp
	bp := f.bp

	// Bias input (6581 only)
	input += f.inputBias

	// State-variable filter core (Chamberlin form)
	bp += w0 * hp
	lp += w0 * bp
	hp = input - lp - f.damping*bp

	// Output saturation (6581 only)
	if curve := f.outputCurve; curve != nil {
		bp = curve(bp)
		hp = curve(hp)
	}

	f.lp = lp
	f.hp = hp
	f.bp = bp
}

func (f *Filter) LowPassOut() float64 {
	return f.lp
}

func (f *Filter) HighPassOut() float64 {
	return f.hp
}

func (f *Filter) BandPassOut() float64 {
	return f.bp
}
