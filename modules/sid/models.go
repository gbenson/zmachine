package sid

// A lot of the numbers and curves in this file is based on a
// (probably) vibe-coded "Go port of browser-based SID synth".
// (See the package comment in sid.go for more on that) but
// the upshot of that is that I don't really know where the
// numbers come from, and _that_ code was so full of cryptic
// comments I didn't even think it was real until I built it
// and sound came out, which is why all the comments in this
// file are kind of head-scratchy. Humanity is fucked!

import "math"

type Model interface {
	Number() int
}

type model interface {
	Model

	// Filter
	FCtoHz(uint) float64
	HzToFC(float64) uint
	InitCutoffTable([]float64, float64)
	DampingForRES(float64) float64
	FilterInputBias() float64
	FilterOutputCurve() func(float64) float64
}

type sid6581 struct{}
type sid8580 struct{}

var (
	model6581 = sid6581{}
	model8580 = sid8580{}

	Model6581 Model = &model6581
	Model8580 Model = &model8580
)

func (m *sid6581) Number() int {
	return 6581
}

func (m *sid8580) Number() int {
	return 8580
}

const (
	min6581cutoff = 30
	max6581cutoff = 12000
	max8580cutoff = 12500
)

func (m *sid6581) FCtoHz(fc uint) float64 {
	x := float64(fc) / 2047.0
	shaped := x * x * (3 - 2*x)
	// I [gbenson] originally assumed the above was maybe something
	// approximated from measurements off a real chip, but it turns
	// out "y = 3x² - 2x³" is the 3rd-order smoothstep function, a
	// sigmoid between (0, 0) and (1, 1) with zero gradient at both
	// ends.  See <https://en.wikipedia.org/wiki/Smoothstep>.
	// It _could_ still be something based on measurements from a real
	// chip, but imo its more likely a feels-right tweak someone added
	// to the "browser-based SID synth" I found it in.  But, I can't
	// really investigate because I don't have that code, if it ever
	// existed at all, just the vibe-coded Go port.
	return min6581cutoff + shaped*(max6581cutoff-min6581cutoff)
}

func (m *sid6581) HzToFC(f float64) uint {
	shaped := (f - min6581cutoff) / (max6581cutoff - min6581cutoff)
	// From <https://en.wikipedia.org/wiki/Smoothstep#Inverse_Smoothstep>,
	// 3rd-order smoothstep function has an analytical inverse, this insanity:
	x := 0.5 - math.Sin(math.Asin(1-2*shaped)/3)
	// XXX this might be a performance killer, try with float32 if it is...
	return uint(x*2047) & 2047
}

func (m *sid8580) FCtoHz(fc uint) float64 {
	return max8580cutoff * float64(fc) / 2047.0
}

func (m *sid8580) HzToFC(f float64) uint {
	return uint(f*2047/max8580cutoff) & 2047
}

func (m *sid6581) InitCutoffTable(table []float64, sampleRate float64) {
	for fc := range 2048 {
		f := m.FCtoHz(uint(fc))
		w0 := 2.0 * math.Sin(math.Pi*math.Min(f, sampleRate*0.45)/sampleRate)
		table[fc] = min(w0, 0.95)
	}
}

func (m *sid8580) InitCutoffTable(table []float64, sampleRate float64) {
	for fc := range 2048 {
		f := m.FCtoHz(uint(fc))
		w0 := 2.0 * math.Sin(math.Pi*f/sampleRate)
		table[fc] = min(w0, 0.95)
	}
}

func (m *sid6581) DampingForRES(res float64) float64 {
	// a) the vibe code calls this "linear 1/Q mapping"
	// b) the 0.06 clip says "res=15 → near-self-oscillation"
	return max(0.06, (15-res)/8)
}

func (m *sid8580) DampingForRES(res float64) float64 {
	// The vibe code calls this "exponential 1/Q resonance"
	// and "Resonance: exponential mapping".
	return math.Pow(2, (4-res)/8)
}

func (m *sid6581) FilterInputBias() float64 {
	return 0.005
}

func (m *sid8580) FilterInputBias() float64 {
	return 0
}

func (m *sid6581) FilterOutputCurve() func(float64) float64 {
	return filterOutputCurve6581
}

func filterOutputCurve6581(v float64) float64 {
	// vibe: "asymmetric saturation for 6581 model"
	// vibe: "Subtle asymmetric warmth — only colors loud signals"
	switch {
	case v > 0.8:
		return 0.8 + (v-0.8)*0.3
	case v < -0.8:
		return -0.8 + (v+0.8)*0.4
	default:
		return v
	}
}

func (m *sid8580) FilterOutputCurve() func(float64) float64 {
	return nil // linear
}
