package main

import (
	"context"
	"os"

	"gbenson.net/go/logger/log"
	"gbenson.net/go/zmachine"
	zm "gbenson.net/go/zmachine/modules"
	zsdl "gbenson.net/go/zmachine/sdl"
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

	m := zmachine.New()
	defer log.LoggedClose(m, "machine")

	m.Source = &generator{}
	m.Sink = &zsdl.AudioSink{}

	ctx = log.DefaultLogger().WithContext(ctx)
	return zmachine.Run(ctx, m)
}

type generator struct {
	pa zm.PhaseAccumulator

	outputLevel float64
}

func (sg *generator) Start(ctx context.Context) error {
	if err := sg.pa.Start(ctx); err != nil {
		return err
	}

	sg.pa.SetFrequency(440)
	sg.outputLevel = 0.125 // approx -18dB; 7 on a 0..10 ↦ -60..0dB volume knob

	return nil
}

func (sg *generator) Generate(ctx context.Context, buf []float32) (int, error) {
	for i := range buf {
		sg.pa.Step()

		output := sg.pa.Phase()*2 - 1
		output *= sg.outputLevel
		buf[i] = float32(output)
	}
	return len(buf), nil
}
