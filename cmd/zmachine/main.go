package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gbenson.net/go/logger/log"
	"gbenson.net/go/zmachine"
	zm "gbenson.net/go/zmachine/modules"
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

	m := zmachine.New()
	ctx = m.WithContext(ctx)

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	m.Source = &generator{}
	m.Sink = &zsdl.AudioSink{}

	return zmachine.Run(ctx, m)
}

type generator struct {
	arp zm.TestArpeggiator
	pa  zm.PhaseAccumulator

	outputLevel float64
}

func (sg *generator) Start(ctx context.Context) error {
	for _, s := range []util.Starter{&sg.arp, &sg.pa} {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}

	sg.outputLevel = 0.125 // approx -18dB; 7 on a 0..10 ↦ -60..0dB volume knob

	return nil
}

func (sg *generator) Generate(ctx context.Context, buf []float32) (int, error) {
	for i := range buf {
		sg.arp.Step()

		sg.pa.SetFrequency(sg.arp.Frequency())
		sg.pa.Step()

		output := sg.pa.Phase()*2 - 1
		output *= sg.outputLevel
		buf[i] = float32(output)
	}
	return len(buf), nil
}
