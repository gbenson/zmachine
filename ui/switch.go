package ui

type Edge int

const (
	FallingEdge Edge = 1 << iota
	RisingEdge
)

func edgeFromClicked(clicked bool) Edge {
	if clicked {
		return RisingEdge
	} else {
		return FallingEdge
	}
}
