package ui

import (
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/conn/v3/display"
	"periph.io/x/devices/v3/ssd1306/image1bit"

	"gbenson.net/go/logger/log"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Renderer interface {
	draw.Image
	SetFont(f font.Face)
	DrawText(x, y int, text string)
}

type renderer struct {
	target     display.Drawer
	fb         *image1bit.VerticalLSB
	fontDrawer font.Drawer
	fontAscent int
}

func newRenderer(target display.Drawer) *renderer {
	if target == nil {
		panic("nil target")
	}

	fb := image1bit.NewVerticalLSB(target.Bounds())
	r := &renderer{
		target: target,
		fb:     fb,
		fontDrawer: font.Drawer{
			Src: &image.Uniform{C: image1bit.On},
			Dst: fb,
		},
	}

	return r
}

func (r *renderer) clear() {
	b := r.fb.Pix
	for i := range b {
		b[i] = 0
	}
}

// present updates the screen with any rendering performed since the
// previous call.
func (r *renderer) present() {
	if err := r.target.Draw(r.Bounds(), r.fb, image.Point{}); err != nil {
		log.Warn().Err(err).Msg("")
	}
}

// Bounds implements [image.Image].
func (r *renderer) Bounds() image.Rectangle {
	return r.fb.Bounds()
}

// ColorModel implements [image.Image].
func (r *renderer) ColorModel() color.Model {
	return r.fb.ColorModel()
}

// At implements [image.Image].
func (r *renderer) At(x, y int) color.Color {
	return r.fb.At(x, y)
}

// Set implements [draw.Image].
func (r *renderer) Set(x, y int, c color.Color) {
	r.fb.Set(x, y, c)
}

func (r *renderer) SetFont(f font.Face) {
	m := f.Metrics()
	r.fontDrawer.Face = f
	r.fontAscent = m.Ascent.Round()
}

func (r *renderer) DrawText(x, y int, text string) {
	r.fontDrawer.Dot = fixed.P(x, y+r.fontAscent)
	r.fontDrawer.DrawString(text)
}
