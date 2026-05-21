package ui

import (
	"math"
	"sync/atomic"
	"time"
)

type encoder struct {
	minusOneOverTC float64

	// Values used by both receive and collect must be safe for
	// concurrent access.
	movement atomic.Int64
	edges    atomic.Uintptr

	// Values used by collect but not receive don't need to worry
	// about concurrency.
	lastV float64   // last non-zero value returned by collectMovement
	lastT time.Time // time when the last non-zero value was returned
}

// setAcceleration sets how much acceleration to apply to this encoder.
// If TC is positive and non-zero, the encoder is treated as having a
// speed, which received movement causes to accelerate, with TC being
// the time constant at which the speed decays.  If TC is zero then no
// acceleration is applied, and collectMovement returns whatever was
// received.  TL;dr larger TC means more acceleration.
func (e *encoder) setAcceleration(tc time.Duration) {
	if tc < 0 {
		panic("cannot be negative")
	} else if tc == 0 {
		e.minusOneOverTC = 0
	} else {
		e.minusOneOverTC = -1 / float64(tc)
	}
}

func (e *encoder) receiveMovement(amount int) {
	e.movement.Add(int64(amount))
}

func (e *encoder) collectMovement() float64 {
	amount := e.movement.Swap(0)
	if amount == 0 {
		return 0
	}

	minusOneOverTC := e.minusOneOverTC
	if minusOneOverTC == 0 {
		return float64(amount) // no acceleration
	}

	v := e.lastV
	t := time.Now()

	if (amount < 0) != math.Signbit(v) {
		// direction change
		v = 0
	} else if lastT := e.lastT; !lastT.IsZero() {
		if dt := t.Sub(lastT); dt > 0 {
			// decay previous speed
			v *= math.Exp(float64(dt) * minusOneOverTC)
		}
	}

	// apply collected movement as acceleration
	v += float64(amount)

	e.lastV = v
	e.lastT = t

	return v
}

func (e *encoder) receiveEdge(clicked bool) {
	e.receiveEdges(edgeFromClicked(clicked))
}

func (e *encoder) receiveEdges(edges Edge) {
	e.edges.Or(uintptr(edges))
}

func (e *encoder) collectEdges() Edge {
	return Edge(e.edges.Swap(0))
}
