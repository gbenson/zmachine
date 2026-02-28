package sdl

// typedef unsigned char Uint8;
// void fillBuffer(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"context"
	"fmt"
	"io"
	"math/bits"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine"
	"github.com/veandco/go-sdl2/sdl"
)

type AudioSink struct {
	DeviceName string

	deviceID     sdl.AudioDeviceID
	deviceOpened atomic.Bool
	log          *logger.Logger
	pinner       runtime.Pinner
	reader       io.Reader
}

// Start implements [machine.Sink].
func (sink *AudioSink) Start(ctx context.Context, r io.Reader) error {
	sink.log = zmachine.Logger(ctx, sink)
	machine := zmachine.FromContext(ctx)

	deviceName := sink.DeviceName
	sampleRate := machine.SampleRate
	maxLatency := machine.MaxLatency

	// Calculate the maximum number of frames we can buffer without
	// exceeding the required maximum latency at the requested sample
	// rate.
	maxBufferFrames := uint(sampleRate) * uint(maxLatency) / uint(time.Second)

	// SDL wants a power of two, so we round down from our maximum.
	// https://wiki.libsdl.org/SDL2/SDL_OpenAudioDevice says "good
	// values seem to range between 512 and 4096 inclusive, depending
	// on the application and CPU speed.  Smaller values reduce
	// latency but can lead to underflow if the application is doing
	// heavy processing and cannot fill the audio buffer in time."
	bufferFrames := 1 << (bits.Len(maxBufferFrames) - 1)

	sink.pinner.Pin(sink)

	desiredSpec := sdl.AudioSpec{
		Freq:     int32(sampleRate),
		Format:   sdl.AUDIO_F32SYS,
		Channels: 1,
		Samples:  uint16(bufferFrames),
		Callback: sdl.AudioCallback(C.fillBuffer),
		UserData: unsafe.Pointer(sink),
	}

	var spec sdl.AudioSpec
	dev, err := sdl.OpenAudioDevice(deviceName, false, &desiredSpec, &spec, 0)
	if err != nil {
		return err
	}
	sink.deviceID = dev
	sink.deviceOpened.Store(true)

	log := sink.log.With().
		Uint32("device_id", uint32(sink.deviceID)).
		Logger()
	sink.log = &log

	log.Info().
		Int32("sample_rate", spec.Freq).
		Uint16("format", uint16(spec.Format)).
		Uint8("channels", spec.Channels).
		Uint16("bufsiz_frames", spec.Samples).
		Uint32("bufsiz_bytes", spec.Size).
		Msg("Using SDL audio")

	if spec.Format != sdl.AUDIO_F32SYS {
		return fmt.Errorf("unexpected sample format 0x%x", spec.Format)
	} else if spec.Channels != 1 {
		return fmt.Errorf("unexpected number of channels (%d)", spec.Channels)
	}

	if uint(spec.Freq) != uint(sampleRate) {
		log.Warn().Int32("sample_rate", spec.Freq).Msg("Unexpected")
	}
	if int(spec.Samples) != bufferFrames {
		log.Warn().Uint16("buffer_size", spec.Samples).Msg("Unexpected")
	}

	sink.reader = r

	sink.trace("Starting")
	defer sink.trace("Started")

	sdl.PauseAudioDevice(sink.deviceID, false)
	return nil
}

// Close implements [io.Closer].
func (sink *AudioSink) Close() error {
	defer sink.pinner.Unpin()
	defer sink.trace("Unpinning")

	if !sink.deviceOpened.Swap(false) {
		return nil
	}

	sink.trace("Closing")
	defer sink.trace("Closed")

	sdl.CloseAudioDevice(sink.deviceID)
	return nil
}

func (sink *AudioSink) trace(msg string) {
	sink.log.Trace().Msg(msg)
}

//export fillBuffer
func fillBuffer(sinkPtr unsafe.Pointer, stream *C.Uint8, length C.int) {
	sink := (*AudioSink)(sinkPtr)

	src := sink.reader
	dst := unsafe.Slice((*byte)(unsafe.Pointer(stream)), length)

	if _, err := io.ReadFull(src, dst); err != nil {
		sink.log.Warn().Err(err).Msg("")
	}
}
