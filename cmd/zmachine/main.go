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
	zm "gbenson.net/go/zmachine/modules"
	"gbenson.net/go/zmachine/modules/sid"
	zsdl "gbenson.net/go/zmachine/sdl"
	"gbenson.net/go/zmachine/util"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Err(err).Msg("Fatal error")
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	defer sdl.Quit()

	logger := log.DefaultLogger()
	ctx = logger.WithContext(ctx)
	lc := util.NewLoggingCloser(ctx)

	m := zmachine.New()
	ctx = m.WithContext(ctx)

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	g := &generator{}
	if err := g.Start(ctx); err != nil {
		return err
	}

	r := zmachine.Open(ctx, g)
	defer lc.Close(r)

	sink := &zsdl.AudioSink{}
	if err := sink.Start(ctx, r); err != nil {
		return err
	}
	defer lc.Close(sink)

	<-ctx.Done()

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

type generator struct {
	arp   zm.TestArpeggiator
	voice zm.Voice
	pa    zm.PhaseAccumulator
	lfo0  zm.PhaseAccumulator
	lfo1  zm.PhaseAccumulator
	lfo2  zm.PhaseAccumulator
	lfo3  zm.PhaseAccumulator
	filt  sid.Filter

	outputLevel float64
}

func (sg *generator) Start(ctx context.Context) error {
	sg.arp.Receiver = &sg.voice
	sg.filt.Model = sid.Model6581

	for _, s := range []util.Starter{
		&sg.voice,
		&sg.arp,
		&sg.pa,
		&sg.lfo0,
		&sg.lfo1,
		&sg.lfo2,
		&sg.lfo3,
		&sg.filt,
	} {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}

	sg.lfo0.SetFrequency(7 * BPM) // lfo1 (cutoff) variation
	//sg.lfo1.SetFrequency(457 * Hz)  // cutoff
	sg.lfo2.SetFrequency(17 * BPM) // res
	sg.lfo3.SetFrequency(29 * BPM) // mode

	sg.filt.SetFC(2000)
	sg.outputLevel = 0.125 // approx -18dB; 7 on a 0..10 ↦ -60..0dB volume knob

	return nil
}

func (sg *generator) Generate(ctx context.Context, buf []float32) (int, error) {
	for i := range buf {
		sg.arp.Step()
		sg.voice.Step()

		sg.pa.SetFrequency(sg.voice.Pitch())
		sg.pa.Step()

		oscmix := sg.pa.Phase().Float64()*2 - 1 // sawtooth

		sg.lfo0.Step()
		sg.lfo1.SetFrequency(Frequency(60/11 + sg.lfo0.Phase()*37))
		sg.lfo1.Step() // cutoff
		sg.lfo2.Step() // resonance
		sg.lfo3.Step() // mode

		sg.filt.SetFrequency(Frequency(500 + 11500*sg.lfo1.Phase()))
		sg.filt.SetResonance(sg.lfo2.Phase())

		sg.filt.SetInput(oscmix)
		sg.filt.Step()

		var output float64
		switch int(sg.lfo1.Phase() * 3) {
		case 1:
			output = sg.filt.BandPassOut()
		case 2:
			output = sg.filt.HighPassOut()
		default:
			output = sg.filt.LowPassOut()
		}

		output *= sg.outputLevel
		buf[i] = float32(output)
	}
	return len(buf), nil
}
