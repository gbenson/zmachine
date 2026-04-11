package ui

import (
	"context"
	"encoding/json"
	"fmt"

	"gbenson.net/go/zmachine/ui/internal/logfollower"
)

type UI struct {
	Display Display
}

func (ui *UI) Start(ctx context.Context) error {
	for _, step := range []func(context.Context) error{
		ui.Display.Start,
	} {
		if err := step(ctx); err != nil {
			defer ui.Stop(ctx)
			return err
		}
	}

	return nil
}

func (ui *UI) Stop(ctx context.Context) {
	defer ui.Display.Stop(ctx)
}

func (ui *UI) ensureDisplay(ctx context.Context) error {
	return ui.Display.Start(ctx)
}

// Follow causes the UI to receive events from the specified source.
func (ui *UI) Follow(es any) {
	switch s := es.(type) {
	case *logfollower.Follower:
		s.AddReceiver(logfollower.ReceiverFunc(ui.onLogRecord))
	default:
		panic(fmt.Sprintf("%T: not implemented", es))
	}
}

// onLogRecord is called whenever a message is logged at info level or higher.
func (ui *UI) onLogRecord(rr *logfollower.Record) {
	var r logRecord
	if err := json.Unmarshal([]byte(rr.Payload), &r); err != nil {
		r.Level = "error"
		r.Component = "ui.LogFollower"
		r.Message = err.Error()
	}
	if msg := r.ShortString(); msg != "" {
		ui.Display.PushMessage(msg)
	}
}
