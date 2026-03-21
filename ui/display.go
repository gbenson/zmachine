package ui

import (
	"context"
	"image"

	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"

	"gbenson.net/go/ssd1305"
	"gbenson.net/go/zmachine"
	"gbenson.net/go/zmachine/machine"
	"gbenson.net/go/zmachine/ui/internal/ssd1305emu"
	"gbenson.net/go/zmachine/util"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Display struct {
	Config *machine.DisplayConfig

	Port   spi.PortCloser
	Device *ssd1305.SSD1305

	shouldClosePort   bool
	shouldCloseDevice bool
}

func (d *Display) Start(ctx context.Context) error {
	log := util.Logger(ctx, d)

	if d.Device != nil {
		log.Debug().Msg("Using supplied device")
		return nil
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
	} {
		if err := step(ctx); err != nil {
			return err
		}
	}

	util.Logger(ctx, d).Info().Stringer("device", d.Device).Msg("Opened")

	// XXX >>>>>
	dev := d.Device
	img := image1bit.NewVerticalLSB(dev.Bounds())
	drawer := font.Drawer{
		Src:  &image.Uniform{C: image1bit.On},
		Dst:  img,
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0, 12),
	}
	drawer.DrawString("Hello world!")
	if err := dev.Draw(dev.Bounds(), img, image.Point{}); err != nil {
		util.Logger(ctx, d).Warn().Err(err).Msg("")
	}
	// <<<<< XXX

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
		d.Port = &ssd1305emu.Port{
			Logger: *util.Logger(ctx, "ssd1305emu"),
		}
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
		Width:    c.SSD1305.Width,
		Height:   c.SSD1305.Height,
		StartCol: c.SSD1305.StartCol,
	}

	if emu, ok := dev.Port.(*ssd1305emu.Port); ok {
		// emulated
		dev.DC = &emu.DC
		dev.RST = &emu.RST
	} else {
		// hardware
		dev.DC = gpioreg.ByName(c.SSD1305.DC)
		dev.RST = gpioreg.ByName(c.SSD1305.RST)
	}

	if err := dev.Open(); err != nil {
		return err
	}
	d.Device = dev
	d.shouldCloseDevice = true

	return nil
}
