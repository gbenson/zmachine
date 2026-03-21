package ssd1305emu

import (
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
)

type PinOut struct {
	name     string
	conn     *spiconn
	hasLevel bool
	level    gpio.Level
}

func (p *PinOut) connect(conn *spiconn, name string) {
	if p.conn != nil {
		panic("already connected")
	}
	p.conn = conn
	p.name = name
	p.hasLevel = false
}

// Function implements [pin.Pin].
func (p *PinOut) Function() string {
	panic("Deprecated: Use PinFunc.Func.")
}

// Halt implements [conn.Resource].
func (p *PinOut) Halt() error {
	panic("ssd1305emu.PinOut.Halt")
}

// Name implements [pin.Pin].
func (p *PinOut) Name() string {
	return p.name
}

// Number implements [pin.Pin].
func (p *PinOut) Number() int {
	panic("ssd1305emu.PinOut.Number")
}

// Out implements [gpio.PinOut].
func (p *PinOut) Out(l gpio.Level) error {
	changed := !p.hasLevel || l != p.level

	p.level = l
	p.hasLevel = true

	if !changed {
		return nil
	}

	return p.conn.port.edge(p)
}

// PWM implements [gpio.PinOut].
func (p *PinOut) PWM(duty gpio.Duty, f physic.Frequency) error {
	panic("ssd1305emu.PinOut.PWM")
}

// String implements [fmt.Stringer].
func (p *PinOut) String() string {
	return "ssd1305emu:" + p.name
}
