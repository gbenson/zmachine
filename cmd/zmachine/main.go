package main

import (
	"context"
	"os"

	"gbenson.net/go/logger/log"
	"gbenson.net/go/zmachine"
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

	ctx = log.DefaultLogger().WithContext(ctx)
	return zmachine.Run(ctx, m)
}
