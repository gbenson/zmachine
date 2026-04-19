package ui

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"gbenson.net/go/logger"
	"gbenson.net/go/microfont"
	"gbenson.net/go/zmachine/util"
)

type systemMenu struct {
	log          *logger.Logger
	bootTimeLine string
	shutdownSelX atomic.Int32
}

func (m *systemMenu) init(ctx context.Context) {
	m.log = util.Logger(ctx, m)
}

// onFirstStep is called just before the audio buffer is filled
// for the first time.
func (m *systemMenu) onFirstStep() {
	d, err := util.Uptime()
	if err != nil {
		m.log.Warn().Err(err).Msg("")
		return
	}

	s := "bootup time: "
	if d < 100*time.Second {
		s += fmt.Sprintf("%.3fs", d.Seconds())
	} else {
		s += m.humanUptime(d)
	}
	m.bootTimeLine = s

	m.log.Debug().Float64("uptime_at_first_step", d.Seconds()).Msg("")
}

func (m *systemMenu) uptimeLine() string {
	d, err := util.Uptime()
	if err != nil {
		return err.Error()
	}
	return "uptime: " + m.humanUptime(d)
}

func (m *systemMenu) humanUptime(d time.Duration) string {
	switch {
	case d < 2*time.Minute:
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	case d < 2*time.Hour:
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%d hours", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days", int(d.Hours()/24))
	}
}

// Update implements [Updatable].
func (m *systemMenu) Update(deltas []int, edges []Edge) {
	const enc = encoderB
	x := m.shutdownSelX.Add(int32(deltas[enc]))
	switch {
	case x < shutdownYesBoxMinX:
	case x > shutdownYesBoxMaxX:
	case edges[enc]&FallingEdge == 0:
	default:
		m.shutdownSystem()
	}
}

func (m *systemMenu) shutdownSystem() {
	ctx := m.log.WithContext(context.Background())
	if err := util.ExecSudo(ctx, "poweroff"); err != nil {
		m.log.Err(err).Msg("")
	}
}

const shutdownYesBoxMinX = 79
const shutdownYesBoxMaxX = 100

// Render implements [Renderable].
func (m *systemMenu) Render(r *Renderer) {
	r.SetFont(microfont.Face04B03B)
	r.DrawText(0, 0, m.uptimeLine())
	r.DrawText(0, 8, m.bootTimeLine)

	r.SetFont(microfont.Face04B08)
	r.DrawText(0, 26, "SHUTDOWN? NO YES")

	const xoff = 3
	x := int(m.shutdownSelX.Load()) - xoff // offset makes it start offscreen
	r.DrawText(x-2, 19, "\u2193")          // -2 makes x be the center of the arrow

	for _, mo := range []struct{ start, limit int }{
		{58, 73},
		{shutdownYesBoxMinX - xoff, shutdownYesBoxMaxX - xoff},
	} {
		if x < mo.start {
			break
		} else if x > mo.limit {
			continue
		}

		fb := r.framebuf
		rowStart := 3 * fb.Stride
		for i := mo.start; i < mo.limit; i++ {
			fb.Pix[rowStart+i] ^= 254
		}

		break
	}
}
