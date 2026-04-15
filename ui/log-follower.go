package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gbenson.net/go/logger"
	"gbenson.net/go/logger/log"
	"gbenson.net/go/microfont"
	"gbenson.net/go/zmachine/midi"
	"gbenson.net/go/zmachine/util"

	"github.com/rs/zerolog"
)

// logFollower is a [Page] that displays logged messages.
type logFollower struct {
	once   sync.Once
	logger *logger.Logger
	lw     zerolog.LevelWriter
	ch     chan string
	wg     sync.WaitGroup
	msgs   []string
	mu     sync.RWMutex
}

// Logger returns a logger that updates the log follower page.
func (f *logFollower) Logger() *logger.Logger {
	f.once.Do(func() {
		f.ch = make(chan string)
		f.wg.Go(func() {
			for s := range f.ch {
				f.receive([]byte(s))
			}
		})

		w := logger.DefaultWriter()
		lw, ok := w.(zerolog.LevelWriter)
		if !ok {
			lw = &zerolog.LevelWriterAdapter{
				Writer: w,
			}
		}
		f.lw = lw

		// We update DefaultLoggerOptions, so we'll see DefaultLogger
		// output if it wasn't already created, but we create our own
		// logger, too, to make sure we at least see our own messages.
		options := log.DefaultLoggerOptions
		options.Writer = f
		l := logger.New(options)
		f.logger = &l
	})
	return f.logger
}

// Close implements [io.Closer].
func (f *logFollower) Close() error {
	if f.ch != nil {
		close(f.ch)
	}
	f.wg.Wait()
	return nil
}

// Write implements [io.Writer].
func (f *logFollower) Write(b []byte) (n int, err error) {
	return 0, util.NotImplementedError("ui.logFollower.Write")
}

// WriteLevel implements [zerolog.LevelWriter].
func (f *logFollower) WriteLevel(l zerolog.Level, b []byte) (n int, err error) {
	if l > zerolog.DebugLevel {
		f.forward(string(b))
	}
	return f.lw.WriteLevel(l, b)
}

// forward protects against writing-to-closed-channel panics.
func (f *logFollower) forward(record string) {
	defer func() { recover() }()
	f.ch <- record
}

// receive receives intercepted records.
func (f *logFollower) receive(b []byte) {
	var r logRecord
	if err := json.Unmarshal(b, &r); err != nil {
		r.Level = "error"
		r.Component = "ui.LogFollower"
		r.Message = err.Error()
	}

	msg := r.ShortString()
	if msg == "" {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.msgs = append(f.msgs, msg)
}

type logRecord struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Component string `json:"comp"`
	Source    string `json:"source"`
}

func (r *logRecord) ShortString() string {
	return shortString(r.Level, r.Message, r.Component, r.Source)
}

func shortString(l, m, c, s string) string {
	// fast path for normal operation
	if l == "info" && m != "" {
		switch c {
		case "":
			return m
		case "midi.Follower":
			switch s {
			case "Midi Through":
				return ""
			case midi.ControlSurfaceName:
				c = "ui.Surface"
			default:
				c = s
			}
		}
		return fmt.Sprintf("%s: %s", c, strings.ToLower(m))
	}

	// non-normal messages
	switch l {
	case "warn":
		l = "warning"
	case "fatal":
		l = "fatal error"
	}

	if c == "" {
		c = "unattributed"
	}

	return fmt.Sprintf("%s %s!", c, l)
}

// Render implements [Renderable].
func (f *logFollower) Render(r *Renderer) {
	r.SetFont(microfont.Face04B03)

	f.mu.RLock()
	defer f.mu.RUnlock()

	msgs := f.msgs
	if offset := len(msgs) - r.Rows(); offset > 0 {
		msgs = msgs[offset:]
	}

	for row, msg := range msgs {
		r.DrawText(0, row*r.FontHeight, msg)
	}
}
