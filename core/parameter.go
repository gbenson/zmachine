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

type Parameters map[string]*Parameter

// Get returns the parameter with the given name.
// It panics if it cannot return a usable parameter.
func (pp Parameters) Get(name string) *Parameter {
	p, found := pp[name]
	switch {
	case !found:
		panic("unknown parameter " + name)
	case p == nil:
		panic("nil parameter " + name)
	case p.Min >= p.Max:
		panic("uninitialized parameter " + name)
	}
	return p
}

// Load atomically loads and returns the value of the parameter.
func (p *Parameter) Load() float64 {
	return math.Float64frombits(p.value.Load())
}

// Store atomically updates the value of the parameter.
func (p *Parameter) Store(v float64) {
	p.value.Store(math.Float64bits(max(p.Min, min(p.Max, v))))
}
