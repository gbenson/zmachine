package main

import (
	"context"
	"os"

	"gbenson.net/go/logger/log"
	"gbenson.net/go/zmachine"
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
}

func (sg *generator) Generate(ctx context.Context, buf []float32) (int, error) {
	m := len(buf)
	n := m / 2
	for i := 0; i < n; i++ {
		buf[i] = 1
	}
	for i := n; i < m; i++ {
		buf[i] = -1
	}
	return len(buf), nil
}
