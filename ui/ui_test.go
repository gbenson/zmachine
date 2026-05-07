package ui

import (
	"testing"

	"gbenson.net/go/zmachine/util/testutil"
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

func (tp *testPage) Update(deltas []float64, edges []Edge) {
	tp.deltas = append(tp.deltas, deltas...)
}

func TestStepUpdate(t *testing.T) {
	var ui UI
	StartForTest(t, &ui)

	var tp testPage
	p := Page(&tp)
	assert.Check(t, ui.currentPage.Swap(&p) == nil)

	assert.Equal(t, ui.stepped.Swap(true), false)

	for i, _ := range ui.surface.encoders {
		if i == int(menuEncoder) {
			continue
		}
		ui.surface.encoders[i].receiveMovement(i*i + 2*i + 1)
	}

	assert.Check(t, tp.deltas == nil)
	ui.Step()
	assert.Equal(t, len(tp.deltas), 4)
	assert.Equal(t, tp.deltas[encoderA], 1.0)
	assert.Equal(t, tp.deltas[encoderB], 4.0)
	assert.Equal(t, tp.deltas[encoderC], 9.0)
	assert.Equal(t, tp.deltas[encoderD], 16.0)
}
