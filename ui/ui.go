package ui

import (
	"context"
	"sync/atomic"
	"time"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
	"periph.io/x/devices/v3/ssd1306/image1bit"
)

type UI struct {
	Display Display
	surface surface
	stepped atomic.Bool

	currentPage  atomic.Pointer[Page]
	pages        []Page
	selectedPage atomic.Uintptr

	loggerPage logFollower
	systemMenu systemMenu
}

// Start implements [Starter].
func (ui *UI) Start(ctx context.Context) error {
	// Display first...
	if ui.Display.Renderer == nil {
		ui.Display.Renderer = ui
	}
	if err := ui.Display.Start(ctx); err != nil {
		return err
	}

	// ...then everything else.
	ui.surface.init(ctx)
	ui.systemMenu.init(ctx)
	ui.AddPage(&ui.systemMenu)
	return nil
}

func (ui *UI) Stop(ctx context.Context) {
	defer util.LoggedClose(ctx, &ui.loggerPage)
	defer ui.Display.Stop(ctx)
}

// Logger returns a logger that updates the log follower page.
// It is safe to call this method and use the returned logger
// at any time, irrespective of whether Start and/or Stop have
// been called.
func (ui *UI) Logger() *logger.Logger {
	return ui.loggerPage.Logger()
}

// Receive implements [zmachine.MIDISink].
func (ui *UI) Receive(msg midi.Message) {
	ui.Display.KeepAlive()
}

// ReceiveFromSurface implements [ControlSink].
func (ui *UI) ReceiveFromSurface(msg midi.Message) {
	ui.surface.Receive(msg)
	ui.Display.KeepAlive()
}

// ControlSurface returns the [core.MIDISink] that interprets
// control change messages from the (hardware) control surface.
func (ui *UI) ControlSurface() core.MIDISink {
	return &ui.surface
}

// AddPage appends a page to the main menu.  It panics if Step
// has been called.
func (ui *UI) AddPage(p Page) {
	if p == nil {
		panic("nil page")
	} else if ui.stepped.Load() {
		// This isn't foolproof, you could race it, but it's
		// intended as a reminder that there's no lock here,
		// you're supposed to add the pages at startup then
		// run the synth and leave them alone.
		panic("UI has been stepped")
	}
	ui.pages = append(ui.pages, p)
}

// CurrentPage returns the currently displayed page.
func (ui *UI) CurrentPage() Page {
	if p := ui.currentPage.Load(); p != nil {
		return *p
	}
	return &ui.loggerPage
}

// Render implements [Renderable].  This is called once per frame
// at the display framerate if a display is configured and enabled.
func (ui *UI) Render(r Renderer) {
	ui.CurrentPage().Render(r)
	if !ui.stepped.Load() {
		ui.renderThrobber(r)
	}
}

// renderThrobber animates a dot on the right-hand edge of the screen.
func (ui *UI) renderThrobber(r Renderer) {
	b := r.Bounds()

	const shift = 4
	mask := uint64(b.Dy()) - 1

	y := (uint64(time.Now().UnixMilli()) >> 4) % mask
	if y > (mask / 2) {
		y = mask - y
	}
	r.Set(b.Max.X-1, int(y), image1bit.On)
}

// Step is called every time the audio buffer is filled.  Note that
// this is almost certainly much faster than the display framerate:
// the default settings (48KHz audio, <10ms latency) will use a
// 256-sample audio buffer, which will cause Step to be called at
// 48,000 ÷ 256 = 187.5 Hz.
func (ui *UI) Step() {
	isFirstStep := !ui.stepped.Swap(true)
	if isFirstStep {
		ui.systemMenu.onFirstStep()
	}

	collected := ui.surface.Scan()

	delta := collected.encoderDeltas[menuEncoder]
	if delta != 0 || isFirstStep {
		index := ui.selectedPage.Add(uintptr(delta))
		count := len(ui.pages)
		if count < 1 {
			return // no pages == no redraws
		}
		page := ui.pages[index%uintptr(count)]
		ui.currentPage.Store(&page)
	} else {
		page := ui.CurrentPage()
		if p, ok := page.(Updatable); ok {
			deltas := collected.encoderDeltas[:encoderD+1]
			edges := collected.encoderEdges[:encoderD+1]
			p.Update(deltas, edges)
		}
	}
}
