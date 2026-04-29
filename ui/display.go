package ui

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"periph.io/x/conn/v3/display"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"

	"gbenson.net/go/ssd1305"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/machine"
	"gbenson.net/go/zmachine/ui/internal/ssd1305emu"
	"gbenson.net/go/zmachine/util"
)

type Display struct {
	Config   *machine.DisplayConfig
	Port     spi.PortCloser
	Device   display.Drawer
	Renderer Renderable

	shouldClosePort   bool
	shouldCloseDevice bool

	wg   sync.WaitGroup
	stop context.CancelFunc

	startTime    time.Time
	lastActivity atomic.Int64
}

// Start implements [Starter].
func (d *Display) Start(ctx context.Context) error {
	if d.Renderer == nil {
		panic("nil renderer")
	}

	d.ensureConfig(ctx)

	log := util.Logger(ctx, d)
	if !d.Enabled() {
		log.Debug().Msg("Display not enabled")
		return nil
	}

	log.Debug().Msg("Starting")
	if err := d.start(ctx); err != nil {
		defer d.Stop(ctx)
		return err
	}

	d.startTime = time.Now()
	d.KeepAlive()

	ctx, d.stop = context.WithCancel(ctx)
	d.wg.Go(func() {
		defer func() { log.Debug().Msg("Stopped") }()

		renderer := newRenderer(d.Device)
		ticker := util.NewTicker(d.Config.FrameRate)
		blanked := false
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				if err := ctx.Err(); !errors.Is(err, context.Canceled) {
					log.Err(err).Msg("")
				}
				return
			}

			if d.shouldBlank() {
				if !blanked {
					if err := d.Device.Halt(); err != nil {
						log.Err(err).Msg("Screen not blanked")
					} else {
						log.Debug().Msg("Screen blanked")
					}
					blanked = true
				}
				continue
			} else if blanked {
				log.Debug().Msg("Screen unblanked")
				blanked = false
			}

			renderer.clear()
			d.Renderer.Render(renderer)
			renderer.present()
		}
	})

	return nil
}

func (d *Display) start(ctx context.Context) error {
	log := util.Logger(ctx, d)
	if d.Device != nil {
		log.Debug().Msg("Using supplied device")
		return nil
	}

	if err := util.StartAll(
		ctx,
		[]Starter{
			StarterFunc(d.ensureDefaults),
			StarterFunc(d.ensureDrivers),
			StarterFunc(d.ensurePort),
			StarterFunc(d.ensureDevice),
		},
	); err != nil {
		return err
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
		if c, ok := d.Device.(io.Closer); ok {
			defer util.LoggedClose(ctx, c)
		}
	}

	if stop := d.stop; stop != nil {
		stop()
	}

	d.wg.Wait()
}

func (d *Display) ensureConfig(ctx context.Context) {
	if d.Config != nil {
		return
	}
	d.Config = &machine.FromContext(ctx).Config.UI.Display
}

func (d *Display) Enabled() bool {
	return d.Config.Enabled
}

func (d *Display) ensureDefaults(ctx context.Context) error {
	switch c := d.Config; c.Type {
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

// KeepAlive prevents the screen from blanking.
func (d *Display) KeepAlive() {
	d.lastActivity.Store(d.now())
}

// shouldBlank reports whether the screen should be blanked.
func (d *Display) shouldBlank() bool {
	blankTime := d.Config.BlankTime
	if blankTime < 1 {
		return false
	}
	return (d.now() - d.lastActivity.Load()) > int64(blankTime)
}

// now returns a monotonic timestamp.
func (d *Display) now() int64 {
	return int64(time.Now().Sub(d.startTime))
}
