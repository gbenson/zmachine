package ssd1305emu

import (
	"bytes"
	"fmt"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"

	"gbenson.net/go/logger"

	"github.com/veandco/go-sdl2/sdl"
)

const MaxWidth = 132
const MaxHeight = 64

type Port struct {
	DC, RST PinOut

	Logger logger.Logger
	regs   Registers
	pixels []byte

	window   *sdl.Window
	renderer *sdl.Renderer
}

type Registers struct {
	Page   int
	Column int
}

// Close implements [io.Closer].
func (p *Port) Close() error {
	if w := p.window; w != nil {
		defer w.Destroy()
	}
	return nil
}

// String implements [fmt.Stringer].
func (p *Port) String() string {
	return "ssd1305emu.Port"
}

// Connect implements [spi.Port].
func (p *Port) Connect(
	f physic.Frequency,
	mode spi.Mode,
	bits int,
) (spi.Conn, error) {
	conn := &spiconn{p}

	p.DC.connect(conn, "DC")
	p.RST.connect(conn, "RST")

	p.pixels = make([]byte, MaxWidth*MaxHeight)
	if err := p.maybeCreateWindow(); err != nil {
		p.Logger.Warn().Err(err).Msg("")
	}

	return conn, nil
}

// LimitSpeed implements [spi.PortCloser].
func (p *Port) LimitSpeed(f physic.Frequency) error {
	panic("ssd1305emu.Port.LimitSpeed")
}

func (p *Port) maybeCreateWindow() error {
	const scale = 4
	w, err := sdl.CreateWindow(
		"ssd1305emu",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		MaxWidth*scale,
		MaxHeight*scale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return err
	}
	p.window = w

	r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}
	p.renderer = r

	if err = p.renderer.SetScale(scale, scale); err != nil {
		return err
	}

	px := p.pixels
	for i, _ := range px {
		px[i] = 128
	}

	p.update()
	return nil
}

// edge is called whenever a pin's level changes.
func (p *Port) edge(pin *PinOut) error {
	if pin != &p.RST || p.RST.level != gpio.Low {
		return nil
	}

	p.Logger.Trace().Msg("Reset")
	return nil
}

// recv is called whenever data is received.
func (p *Port) recv(b []byte) error {
	if p.RST.level == gpio.Low {
		return nil
	} else if p.DC.level == gpio.Low {
		r := p.regs
		if err := r.eval(b); err != nil {
			p.Logger.Trace().
				Hex("data", b).
				//Err(err).
				Msg("Eval")
			return nil
		}
		p.regs = r
		return nil
	} else {
		return p.store(b)
	}
}

// eval interprets commands.
func (r *Registers) eval(b []byte) error {
	br := bytes.NewReader(b)
	for br.Len() > 0 {
		cmd, _ := br.ReadByte()

		switch {
		default:
			return fmt.Errorf("ssh1305emu: 0x%02x: not implemented", cmd)

		case cmd < 0x10:
			// 10.1.1 Set Lower Column Start Address for Page Addressing Mode
			r.Column &= 0xf0
			r.Column |= int(cmd)

		case cmd < 0x20:
			// 10.1.2 Set Higher Column Start Address for Page Addressing Mode
			r.Column &= 15
			r.Column |= ((int(cmd) & 15) << 4)

		case (cmd & 0xf8) == 0xb0:
			// 10.1.20 Set Page Start Address for Page Addressing Mode
			r.Page = int(cmd) & 7
		}
	}

	return nil
}

// store updates GDDRAM.
func (p *Port) store(b []byte) error {
	r := &p.regs
	y0 := r.Page * 8
	x := r.Column
	if chk := x + len(b); chk > MaxWidth {
		return fmt.Errorf("ssh1305emu: %d > %d", chk, MaxWidth)
	}
	px := p.pixels
	for _, bits := range b {
		for y := range 8 {
			var v uint8
			if (bits & (1 << y)) != 0 {
				v = 255
			}
			px[(y+y0)*MaxWidth+x] = v
		}
		x++
	}
	r.Column = x

	p.update()
	return nil
}

// update updates the window
func (p *Port) update() {
	r := p.renderer
	if r == nil {
		return
	}

	px := p.pixels
	for y := 0; y < MaxHeight; y++ {
		for x := 0; x < MaxWidth; x++ {
			v := px[y*MaxWidth+x]
			r.SetDrawColor(v, v, v, 255)
			r.DrawPoint(int32(x), int32(y))
		}
	}

	r.Present()
}
