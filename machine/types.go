package machine

import (
	"context"
	"io"
)

// An AudioSource is a component that generates samples.
type AudioSource interface {
	// Generate stores up to len(buf) samples into buf.
	Generate(ctx context.Context, buf []float32) (n int, err error)
}

// An AudioSink is a component that consumes samples.
type AudioSink interface {
	// Start causes the sink to consume samples until cancelled.
	Start(ctx context.Context, r io.Reader) error
}
