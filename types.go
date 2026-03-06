package zmachine

import (
	"context"
	"io"
)

// An AudioSink is a component that consumes samples.
type AudioSink interface {
	// Start causes the sink to consume samples until cancelled.
	Start(ctx context.Context, r io.Reader) error
}
