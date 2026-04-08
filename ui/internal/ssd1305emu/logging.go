package ssd1305emu

import (
	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/util"
)

type destroyer interface {
	Destroy() error
}

// loggedDestroy logs a call to the Destroy method of an [destroyer].
func loggedDestroy(l *logger.Logger, d destroyer) {
	if d == nil {
		return
	}

	log := l.With().Str("comp", util.ComponentName(d)).Logger()

	log.Debug().Msg("Closing")
	if err := d.Destroy(); err != nil {
		log.Warn().Err(err).Msg("Destroy failed")
	} else {
		log.Debug().Msg("Closed")
	}
}
