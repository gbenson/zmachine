package util

import "context"

// A Starter is a component that must be started before use.
type Starter interface {
	// Start causes a component to run until cancelled.
	Start(ctx context.Context) error
}
