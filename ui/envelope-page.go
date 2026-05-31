package ui

import (
	"strconv"

	"gbenson.net/go/microfont"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/surface"
)

type envelopePage struct {
	Title  string
	Params [4]parameter
}

func NewEnvelopePage(title string, e Parameterized) Page {
	page := &envelopePage{Title: title}
	params := e.Parameters()
	for i, name := range []string{"attack", "decay", "sustain", "release"} {
		value := params.Get(name)

		var curve parameterCurve
		switch name {
		case "sustain":
			curve = newLinearParameterCurve(value, 128)
		default:
			curve = newExponentialParameterCurve(value, 512)
		}

		page.Params[i] = parameter{name, value, curve}
	}
	return page
}

// Render implements [Renderable].
func (page *envelopePage) Render(r Renderer) {
	r.SetFont(microfont.Face04B08)
	r.DrawText(0, 0, page.Title)

	r.SetFont(microfont.Face04B03B)
	for i, p := range page.Params {
		x := i * 128 / 4
		nudge := 0

		v := p.Value.Load()
		prec := 4
		if v >= 10 {
			prec = 3
		} else if v < 0 {
			nudge = -2
		}
		s := strconv.FormatFloat(v, 'f', prec, 64)
		r.DrawText(x+nudge, 16, s)

		if i == 2 { // sustain
			nudge = -3
		}
		r.DrawText(x+nudge, 26, p.Name)
	}
}

// Update implements [Updatable].
func (page *envelopePage) Update(s *surface.State) {
	for i, p := range page.Params {
		p.Update(s.Encoders[i].Delta)
	}
}
