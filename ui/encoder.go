package ui

import "sync/atomic"

type encoder struct {
	movement atomic.Int64
	edges    atomic.Uintptr
}

func (e *encoder) receiveMovement(amount int) {
	e.movement.Add(int64(amount))
}

func (e *encoder) collectMovement() int {
	return int(e.movement.Swap(0))
}

func (e *encoder) receiveEdges(edges Edge) {
	e.edges.Or(uintptr(edges))
}

func (e *encoder) collectEdges() Edge {
	return Edge(e.edges.Swap(0))
}
