package ui

import (
	"testing"

	"gbenson.net/go/zmachine/surface"
	"gbenson.net/go/zmachine/util/testutil"
	"gitlab.com/gomidi/midi/v2"
	"gotest.tools/v3/assert"
)

var (
	NearlyEqual  = testutil.NearlyEqual
	StartForTest = testutil.StartForTest
)

func TestInitialPage(t *testing.T) {
	var ui UI
	assert.Equal(t, ui.CurrentPage(), &ui.loggerPage)
}

type testPage struct {
	deltas []float64
}

func (tp *testPage) Render(r Renderer) {
	panic("should not call")
}

func (tp *testPage) Update(s *surface.State) {
	for _, e := range s.Encoders {
		tp.deltas = append(tp.deltas, e.Delta)
	}
}

func TestStepUpdate(t *testing.T) {
	var ui UI
	StartForTest(t, &ui)

	var tp testPage
	p := Page(&tp)
	assert.Check(t, ui.currentPage.Swap(&p) == nil)

	assert.Equal(t, ui.stepped.Swap(true), false)

	for i := range uint8(4) {
		ui.ReceiveFromSurface(
			midi.ControlChange(
				i, //channel, should be ignored
				midi.GeneralPurposeSlider1+i,
				i*i+2*i+65,
			),
		)
	}

	assert.Check(t, tp.deltas == nil)
	ui.Step()
	assert.Equal(t, len(tp.deltas), int(surface.NumEncoders))
	assert.Equal(t, tp.deltas[surface.EncoderA], 1.0)
	assert.Equal(t, tp.deltas[surface.EncoderB], 4.0)
	assert.Equal(t, tp.deltas[surface.EncoderC], 9.0)
	assert.Equal(t, tp.deltas[surface.EncoderD], 16.0)
}
