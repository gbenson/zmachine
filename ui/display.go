package ui

import (
	"context"
	"image"
	"sync"

	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"

	"gbenson.net/go/ssd1305"
	"gbenson.net/go/zmachine"
	"gbenson.net/go/zmachine/machine"
	"gbenson.net/go/zmachine/util"
)

type Display struct {
	Config *machine.DisplayConfig

	Port   spi.PortCloser
	Device *ssd1305.SSD1305

	shouldClosePort   bool
	shouldCloseDevice bool

	Renderer *Renderer
	Messages []string

	pmMu sync.Mutex
}

func (d *Display) Start(ctx context.Context) error {
	log := util.Logger(ctx, d)

	if d.Renderer != nil {
		log.Debug().Msg("Using supplied renderer")
		return nil
	} else if d.Device != nil {
		log.Debug().Msg("Using supplied device")
		return d.ensureRenderer(ctx)
	}

	c := d.Config
	if c == nil {
		c = &zmachine.FromContext(ctx).Config.UI.Display
	}
	if !c.Enabled {
		log.Debug().Str("reason", "not enabled").Msg("Skipping")
		return nil
	}
	d.Config = c

	for _, step := range []func(context.Context) error{
		d.ensureConfig,
		d.ensureDrivers,
		d.ensurePort,
		d.ensureDevice,
		d.ensureRenderer,
	} {
		if err := step(ctx); err != nil {
			return err
		}
	}

	log.Info().Stringer("device", d.Device).Msg("Opened")
	return nil
}

func (d *Display) Stop(ctx context.Context) {
	if d.shouldClosePort {
		defer func() { d.shouldClosePort = false }()
		defer util.LoggedClose(ctx, d.Port)
	}

	if d.shouldCloseDevice {
		defer func() { d.shouldCloseDevice = false }()
		defer util.LoggedClose(ctx, d.Device)
	}
}

func (d *Display) ensureConfig(ctx context.Context) error {
	c := d.Config
	switch c.Type {
	default:
		return util.NotImplementedError("ui.display.type=" + c.Type)

	case "waveshare_2in23_oled":
		if c.Driver == "" {
			c.Driver = "ssd1305"
		}
		if c.Port == "" {
			c.Port = "spi"
		}
		c.SSD1305.ApplyDefaults(
			&machine.SSD1305Config{
				DC:       "GPIO24",
				RST:      "GPIO25",
				Width:    128,
				Height:   32,
				StartCol: 4,
			},
		)
	}

	return nil
}

func (d *Display) ensureDrivers(ctx context.Context) error {
	_, err := host.Init()
	return err
}

func (d *Display) ensurePort(ctx context.Context) error {
	if d.Port != nil {
		return nil
	}
	defer func() {
		if port := d.Port; port != nil {
			d.shouldClosePort = true
		}
	}()

	c := d.Config
	switch c.Port {
	default:
		return util.NotImplementedError("ui.display.port=" + c.Port)

	case "spi":
		port, err := spireg.Open("")
		if err != nil {
			return err
		}
		d.Port = port

	case "ssd1305emu":
		port, err := ssd1305emu.NewSPI(ctx)
		if err != nil {
			return err
		}
		d.Port = port
	}

	return nil
}

func (d *Display) ensureDevice(ctx context.Context) error {
	c := d.Config

	if c.Driver != "ssd1305" {
		return util.NotImplementedError("ui.display.driver=" + c.Driver)
	}

	dev := &ssd1305.SSD1305{
		Port:     d.Port,
		DC:       gpioreg.ByName(c.SSD1305.DC),
		RST:      gpioreg.ByName(c.SSD1305.RST),
		Width:    c.SSD1305.Width,
		Height:   c.SSD1305.Height,
		StartCol: c.SSD1305.StartCol,
	}

	if err := dev.Open(); err != nil {
		return err
	}
	d.Device = dev
	d.shouldCloseDevice = true

	return nil
}

func (d *Display) ensureRenderer(ctx context.Context) error {
	d.Renderer = NewRenderer(d.Device)
	return nil
}

func (d *Display) PushMessage(msg string) {
	d.pmMu.Lock()
	defer d.pmMu.Unlock()

	msgs := append(d.Messages, msg)
	d.Messages = msgs

	r := d.Renderer
	if r == nil {
		return
	}

	if offset := len(msgs) - r.Rows(); offset > 0 {
		msgs = msgs[offset:]
	}

	r.Clear()
	for row, msg := range msgs {
		r.DrawText(image.Point{0, row * r.FontHeight}, msg)
	}
	r.Present()
}
