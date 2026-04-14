//go:build amd64

package ui

import (
	"context"
	"errors"

	impl "gbenson.net/go/zmachine/ui/internal/ssd1305emu"
	"periph.io/x/conn/v3/spi"
)

type ssd1305emuT struct{}

var ssd1305emu = &ssd1305emuT{}

func (ui *UI) maybeHookEmulation(ctx context.Context) context.Context {
	return context.WithValue(ctx, ssd1305emu, ui)
}

func (e *ssd1305emuT) NewSPI(ctx context.Context) (spi.PortCloser, error) {
	ui, _ := ctx.Value(e).(*UI)
	if ui == nil {
		return nil, errors.New("ssd1305emu: not hooked")
	}

	c := ui.Display.Config
	if c == nil {
		return nil, errors.New("ssd1305emu: not configured")
	}

	p, err := impl.NewSPI(ctx)
	if err != nil {
		return nil, err
	}

	c.SSD1305.DC = p.DC().Name()
	c.SSD1305.RST = p.RST().Name()

	return p, nil
}
