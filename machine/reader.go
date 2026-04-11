package machine

import (
	"context"
	"fmt"
	"io"
	"sync"
	"unsafe"

	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
)

type reader struct {
	ctx     context.Context
	source  AudioSource
	metrics readerMetrics
	onRead  func()
}

// NewReader returns a new [io.Reader] reading from src.
func NewReader(ctx context.Context, src AudioSource) io.ReadCloser {
	if ctx == nil {
		panic("nil context")
	} else if src == nil {
		panic("nil source")
	}

	return &reader{
		ctx:    ctx,
		source: src,
		onRead: sync.OnceFunc(func() {
			util.Logger(ctx, "zmachine").Info().Msg("Running")
		}),
	}
}

// Read implements [io.Reader].
func (r *reader) Read(p []byte) (n int, err error) {
	r.metrics.setWorking()
	r.onRead()
	defer r.metrics.setIdle()

	sizeBytes := len(p)
	if sizeBytes < 1 || (sizeBytes&3) != 0 {
		return 0, fmt.Errorf("unexpected buffer size %d", sizeBytes)
	}
	numSamples := sizeBytes / 4

	ptr := unsafe.Pointer(unsafe.SliceData(p))
	buf := unsafe.Slice((*float32)(ptr), numSamples)

	numSamples, err = r.source.Generate(r.ctx, buf)
	return numSamples * 4, err
}

// Close implements [io.Closer].
func (r *reader) Close() error {
	defer func() { r.metrics.logReport(r.ctx) }()
	return nil
}
