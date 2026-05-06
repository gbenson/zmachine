package core

import (
	"math"
	"sync/atomic"
)

type Parameter struct {
	Min   float64
	Max   float64
	value atomic.Uint64
}

// Load atomically loads and returns the value of the parameter.
func (p *Parameter) Load() float64 {
	return math.Float64frombits(p.value.Load())
}

// Store atomically updates the value of the parameter.
func (p *Parameter) Store(v float64) {
	p.value.Store(math.Float64bits(max(p.Min, min(p.Max, v))))
}
