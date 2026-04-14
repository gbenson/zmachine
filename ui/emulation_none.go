//go:build !amd64

package ui

import "context"

func (ui *UI) maybeHookEmulation(ctx context.Context) context.Context {
	return ctx
}

var ssd1305emu = &ssd1305emuT{}

func (e *ssd1305emuT) NewSPI(ctx context.Context) (spi.PortCloser, error) {
	return nil, errors.New("ssd1305emu: not available")
}
