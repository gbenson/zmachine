package ui

import "context"

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
