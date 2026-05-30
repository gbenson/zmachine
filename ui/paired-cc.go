package ui

import (
	"context"
	"sync"
	"sync/atomic"

	"gbenson.net/go/zmachine/util"
)

type pairedCC struct {
	wg sync.WaitGroup
	ch chan uintptr
	v  atomic.Uintptr
}

const isMSB = 1 << 15

// Start implements [Starter].
func (cc *pairedCC) Start(ctx context.Context) error {
	log := util.Logger(ctx, "volume.Follower")

	cc.ch = make(chan uintptr)

	log.Debug().Msg("Starting")
	cc.wg.Go(func() {
		defer func() { log.Debug().Msg("Stopped") }()

		// surface sends msb, lsb
		var msb uintptr
		for v := range cc.ch {
			if v > 127 {
				msb = v & (127 << 7)
			} else {
				cc.v.Store(uintptr(msb | v))
			}
		}
	})

	return nil
}

func (cc *pairedCC) Stop(ctx context.Context) {
	if ch := cc.ch; ch != nil {
		close(ch)
	}
	cc.wg.Wait()
}

func (cc *pairedCC) receiveMSB(msb uint8) {
	cc.ch <- (uintptr(msb) << 7) | isMSB
}

func (cc *pairedCC) receiveLSB(lsb uint8) {
	cc.ch <- uintptr(lsb)
}

func (cc *pairedCC) Load() int {
	return int(cc.v.Load())
}
