package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"gbenson.net/go/logger/log"
	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/midi"
	zm "gbenson.net/go/zmachine/modules"
	"gbenson.net/go/zmachine/modules/sid"
	zsdl "gbenson.net/go/zmachine/sdl"
	zui "gbenson.net/go/zmachine/ui"
	"gbenson.net/go/zmachine/util"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Err(err).Msg("Fatal error")
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Set up logging before anything else.
	lw := log.DefaultWriter()
	lf := zui.NewLogFollower(lw)
	log.DefaultLoggerOptions.Writer = lf

	logger := log.DefaultLogger()

	logger.Info().Str("comp", "zmachine").Msg("Starting")
	defer func() { logger.Info().Str("comp", "zmachine").Msg("Stopped") }()

	ctx = logger.WithContext(ctx)
	lc := util.NewLoggingCloser(ctx)
	defer lc.Close(lf)

	// Create the machine, and read any config file.
	m := zmachine.New()
	ctx = m.WithContext(ctx)

	if err := m.Config.Read(); err != nil {
		return err
	} else if f := m.Config.Filename; f != "" {
		util.Logger(ctx, m.Config).Info().Str("file", f).Msg("Loaded")
	}

	// Set up signal handling before doing anything needing cleanup.
	// Everything after this point will have its context canceled on
	// any listed signal (or if this function returns.)
	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Start the UI so we can log progress while everything initializes.
	ui := &zui.UI{}
	ui.Follow(lf)
	if err := ui.Start(ctx); err != nil {
		return err
	}
	defer ui.Stop(ctx)

	if sdl.WasInit(sdl.INIT_AUDIO) == 0 {
		log := util.Logger(ctx, "sdl.Audio")
		log.Debug().Msg("Starting")

		if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err != nil {
			return err
		}
		defer sdl.QuitSubSystem(sdl.INIT_AUDIO)

		log.Info().Msg("Started")
	}

	drv, err := rtmididrv.New()
	if err != nil {
		return err
	}
	defer lc.Close(drv)

	g := &generator{}
	if err := g.Start(ctx); err != nil {
		return err
	}

	f := &midi.Follower{
		Driver:   drv,
		Receiver: &g.voice,
	}
	if err := f.Start(ctx); err != nil {
		return err
	}
	defer f.Stop(ctx)

	r := zmachine.Open(ctx, g)
	defer lc.Close(r)

	sink := &zsdl.AudioSink{}
	if err := sink.Start(ctx, r); err != nil {
		return err
	}
	defer lc.Close(sink)

	logger.Info().Msg("Startup complete")
	<-ctx.Done()

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

type generator struct {
	//arp   zm.TestArpeggiator
	voice zm.Voice
	osc1  zm.PhaseAccumulator
	lfo1  zm.PhaseAccumulator
	lfo2  zm.PhaseAccumulator
	filt  sid.Filter

	osc1shaper Shaper
	lfo1shaper Shaper
	lfo2shaper Shaper

	outputLevel Fraction
}

// XXX add a Stopper type to match util.Starter, then, have Start
// build a list of started components that need stopping; implement
// generator.stop; and defer call it above (i.e. in run) and below
// (in Start, if any starter fails). then, make generator.in not a
// pointer, make generator.Start start it (and so also make gen.stop
// stop it), and remove the above midi.Follower.Start.

func (sg *generator) Start(ctx context.Context) error {
	sg.filt.Model = sid.Model6581

	for _, s := range []util.Starter{
		&sg.voice,
		//&sg.arp,
		&sg.osc1,
		&sg.lfo1,
		&sg.lfo2,
		&sg.filt,
	} {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}

	sg.osc1shaper = zm.RisingSawShaper

	// LFO1 modulates cutoff
	sg.lfo1.SetFrequency(11 * BPM)
	sg.lfo1shaper = zm.SineShaper

	// LFO2 modulates resonance
	sg.lfo2.SetFrequency(97 * BPM)
	sg.lfo2shaper = zm.TriangleShaper

	sg.filt.SetFC(2000)

	sg.outputLevel = 0.125 // approx -18dB; 7 on a 0..10 ↦ -60..0dB volume knob

	return nil
}

func (sg *generator) Generate(ctx context.Context, buf []float32) (int, error) {
	for i := range buf {
		//sg.arp.Step()
		sg.voice.Step()

		sg.osc1.SetFrequency(sg.voice.Pitch())
		sg.osc1.Step()

		oscmix := sg.osc1shaper.Sample(sg.osc1.Phase())

		// LFO1 modulates cutoff
		sg.lfo1.Step()
		sg.lfo2.Step()

		lfo1val := sg.lfo1shaper.Fraction(sg.lfo1.Phase())
		lfo2val := sg.lfo2shaper.Fraction(sg.lfo2.Phase())

		sg.filt.SetFrequency(Frequency(500 + 11000*lfo1val.Float64()))
		sg.filt.SetResonance(lfo2val)

		sg.filt.SetInput(oscmix)
		sg.filt.Step()

		output := sg.filt.LowPassOut()

		output *= Sample(sg.outputLevel)
		buf[i] = float32(output)
	}
	return len(buf), nil
}
