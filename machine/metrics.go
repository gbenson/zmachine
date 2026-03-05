package machine

import (
	"context"
	"fmt"
	"math"
	"time"

	"gbenson.net/go/logger"
)

type readerMetrics struct {
	startWork time.Time
	startIdle time.Time
	workTime  time.Duration
	totalTime time.Duration
	numReads  int
}

func (m *readerMetrics) setWorking() {
	m.startWork = time.Now()
}

func (m *readerMetrics) setIdle() {
	startIdle := time.Now()

	m.workTime += startIdle.Sub(m.startWork)
	m.numReads++

	if lastStartIdle := m.startIdle; !lastStartIdle.IsZero() {
		m.totalTime += startIdle.Sub(lastStartIdle)
	}

	m.startIdle = startIdle
}

func (m *readerMetrics) logReport(ctx context.Context) {
	if m.numReads < 1 {
		return
	}
	workTime := m.workTime.Nanoseconds()
	totalTime := m.totalTime.Nanoseconds()

	loadFactor := float64(workTime) / float64(totalTime)
	utilization := formatPercent(loadFactor)

	logger.Ctx(ctx).Debug().
		Int("num_calls", m.numReads).
		Int64("time_active_ns", workTime).
		Int64("time_total_ns", totalTime).
		Str("utilization", utilization).
		Msg("")
}

func formatPercent(v float64) string {
	v *= 100
	n := 1 - int(math.Floor(math.Log10(v)))
	if n < 0 {
		n = 0
	}
	f := fmt.Sprintf("%%.%df%%%%", n)
	return fmt.Sprintf(f, v)
}
