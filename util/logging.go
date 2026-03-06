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
func LoggedClose(ctx context.Context, c io.Closer) {
	NewLoggingCloser(ctx).Close(c)
}

// A LoggingCloser wraps calls to [io.Closer.Close], logging errors as
// warnings.  This avoids silently dropping errors, as would otherwise
// happen if the Close method was directly invoked by defer.
type LoggingCloser struct {
	ctx context.Context
}

// NewLoggingCloser creates and initializes a new [LoggingCloser].
func NewLoggingCloser(ctx context.Context) *LoggingCloser {
	if ctx == nil {
		panic("nil context")
	}
	return &LoggingCloser{ctx}
}

// Close logs a call to the Close method of an [io.Closer].
func (lc *LoggingCloser) Close(c io.Closer) {
	log := Logger(lc.ctx, c)
	log.Debug().Msg("Closing")
	if err := c.Close(); err != nil {
		log.Warn().Err(err).Msg("Close failed")
	} else {
		log.Debug().Msg("Closed")
	}
}
