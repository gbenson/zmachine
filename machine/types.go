package machine

import "context"

// An AudioSource is a component that generates samples.
type AudioSource interface {
	// Generate stores up to len(buf) samples into buf.
	Generate(ctx context.Context, buf []float32) (n int, err error)
}
