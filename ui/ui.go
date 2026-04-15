package ui

import (
	"context"
	"sync/atomic"
	"time"

	"gbenson.net/go/logger"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
	"periph.io/x/devices/v3/ssd1306/image1bit"
)

type UI struct {
	Display Display
	surface surface

	loggerPage  logFollower
	currentPage atomic.Pointer[Page]
}

func (ui *UI) Start(ctx context.Context) error {
	// Display first...
	if err := ui.Display.Start(ctx, ui); err != nil {
		return err
	}

	// ...then everything else.
	ui.surface.init(ctx)
	return nil
}

func (ui *UI) Stop(ctx context.Context) {
	defer util.LoggedClose(ctx, &ui.loggerPage)
	defer ui.Display.Stop(ctx)
}

// Logger returns a logger that updates the log follower page.
// It is safe to call this method and use the returned logger
// at any time, irrespective of whether [Start] and/or [Stop]
// have been called.
func (ui *UI) Logger() *logger.Logger {
	return ui.loggerPage.Logger()
}

// ControlSurface returns the [MIDISink] that interprets control
// change messages from the (hardware) control surface.
func (ui *UI) ControlSurface() MIDISink {
	return &ui.surface
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
func (ui *UI) Render(r *Renderer) {
	r.Clear()
	ui.CurrentPage().Render(r)
	ui.renderThrobber(r)
	r.Present()
}

// renderThrobber animates a dot on the right-hand edge of the screen.
func (ui *UI) renderThrobber(r *Renderer) {
	const shift = 4
	mask := uint64(r.Height*2) - 1

	y := (uint64(time.Now().UnixMilli()) >> shift) & mask
	if y > (mask / 2) {
		y = mask - y
	}
	r.framebuf.Set(r.Width-1, int(y), image1bit.On)
}
