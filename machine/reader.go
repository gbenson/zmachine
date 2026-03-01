package machine

import (
	"context"
	"fmt"
	"unsafe"
)

type reader struct {
	ctx     context.Context
	source  AudioSource
	metrics readerMetrics
}

// Read implements [io.Reader].
func (r *reader) Read(p []byte) (n int, err error) {
	r.metrics.setWorking()
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
