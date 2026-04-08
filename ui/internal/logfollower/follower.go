package logfollower

import (
	"io"
	"sync"

	"github.com/rs/zerolog"
)

type Follower struct {
	lw zerolog.LevelWriter
	ch chan *Record
	mu sync.RWMutex
	wg sync.WaitGroup

	receivers []Receiver
}

type Record struct {
	Level   zerolog.Level
	Payload string
}

// New creates and initializes a new [Follower].
func New(w io.Writer) *Follower {
	lw, ok := w.(zerolog.LevelWriter)
	if !ok {
		lw = &zerolog.LevelWriterAdapter{
			Writer: w,
		}
	}

	f := &Follower{
		lw: lw,
		ch: make(chan *Record),
	}
	f.wg.Go(f.follow)
	return f
}

// Close implements [io.Closer].
func (f *Follower) Close() error {
	close(f.ch)
	f.wg.Wait()
	return nil
}

// AddReceiver adds the given [Receiver].
func (f *Follower) AddReceiver(r Receiver) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.receivers = append(f.receivers, r)
}

// Write implements [io.Writer].
func (f *Follower) Write(b []byte) (n int, err error) {
	panic("should not call")
}

// WriteLevel implements [zerolog.LevelWriter].
func (f *Follower) WriteLevel(l zerolog.Level, b []byte) (n int, err error) {
	if l > zerolog.DebugLevel {
		f.receive(&Record{l, string(b)})
	}
	return f.lw.Write(b)
}

// receive protects against writing-to-closed-channel panics.
func (f *Follower) receive(r *Record) {
	defer func() { recover() }()
	f.ch <- r
}

// follow forwards received records.
func (f *Follower) follow() {
	for r := range f.ch {
		f.forward(r)
	}
}

// forward forwards the given record.
func (f *Follower) forward(r *Record) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, receiver := range f.receivers {
		receiver.Receive(r)
	}
}
