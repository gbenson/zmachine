package ui

import (
	"image"

	"periph.io/x/conn/v3/display"
	"periph.io/x/devices/v3/ssd1306/image1bit"

	"gbenson.net/go/logger/log"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Renderer struct {
	target     display.Drawer
	framebuf   *image1bit.VerticalLSB
	fontDrawer font.Drawer
	fontAscent int

	Width, Height int
	FontHeight    int
}

func NewRenderer(target display.Drawer) *Renderer {
	if target == nil {
		panic("nil target")
	}

	bounds := target.Bounds()
	fb := image1bit.NewVerticalLSB(bounds)

	r := &Renderer{
		target:   target,
		framebuf: fb,
		fontDrawer: font.Drawer{
			Src: &image.Uniform{C: image1bit.On},
			Dst: fb,
		},
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}

	return r
}

func (r *Renderer) SetFont(f font.Face) {
	m := f.Metrics()
	r.fontDrawer.Face = f
	r.fontAscent = m.Ascent.Round()
	r.FontHeight = m.Height.Round()
}

// Return the number of rows of text that may be displayed with the
// current font.
func (r *Renderer) Rows() int {
	return r.Height / r.FontHeight
}

func (r *Renderer) Clear() {
	b := r.framebuf.Pix
	for i := range b {
		b[i] = 0
	}
}

func (r *Renderer) DrawText(x, y int, text string) {
	r.fontDrawer.Dot = fixed.P(x, y+r.fontAscent)
	r.fontDrawer.DrawString(text)
}

func (r *Renderer) Present() {
	fb := r.framebuf
	if err := r.target.Draw(fb.Rect, fb, image.Point{}); err != nil {
		log.Warn().Err(err).Msg("")
	}
}
