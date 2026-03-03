package util

import (
	"context"
	"io"

	"gbenson.net/go/logger"
)

// ComponentLogger returns a [logger.Logger] suitable for comp.
func ComponentLogger(ctx context.Context, comp any) *logger.Logger {
	l := logger.Ctx(ctx).With().
		Str("comp", ComponentName(comp)).
		Logger()
	return &l
}

// LoggedStart logs a call to the Start method of a [Starter].
func LoggedStart(ctx context.Context, s Starter) (err error) {
	log := ComponentLogger(ctx, s)
	log.Debug().Msg("Starting")
	if err = s.Start(ctx); err == nil {
		log.Debug().Msg("Started")
	}
	return
}

// LoggedClose logs a call to the Close method of an [io.Closer].
func LoggedClose(ctx context.Context, c io.Closer) (err error) {
	log := ComponentLogger(ctx, c)
	log.Debug().Msg("Closing")
	if err = c.Close(); err == nil {
		log.Debug().Msg("Closed")
	}
	return
}

// DeferableLoggedClose logs a call to the Close method of an
// [io.Closer], logging any resulting error with warning level.
func DeferableLoggedClose(ctx context.Context, c io.Closer) {
	if err := LoggedClose(ctx, c); err != nil {
		log := ComponentLogger(ctx, c)
		log.Warn().Err(err).Msg("")
	}
}
