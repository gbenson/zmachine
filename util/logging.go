package util

import (
	"context"
	"io"

	"gbenson.net/go/logger"
)

// Logger returns a [logger.Logger] suitable for comp.
func Logger(ctx context.Context, comp any) *logger.Logger {
	l := logger.Ctx(ctx).With().
		Str("comp", ComponentName(comp)).
		Logger()
	return &l
}

// LoggedClose logs a call to the Close method of an [io.Closer].
func LoggedClose(ctx context.Context, c io.Closer) (err error) {
	log := Logger(ctx, c)
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
		log := Logger(ctx, c)
		log.Warn().Err(err).Msg("")
	}
}
