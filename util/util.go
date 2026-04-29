package util

import (
	"context"
	"fmt"
	"strings"

	. "gbenson.net/go/zmachine/core"
)

func ComponentName(comp any) string {
	if name, ok := comp.(string); ok {
		return name
	}
	name := fmt.Sprintf("%T", comp)    // *package.Type
	name = strings.TrimLeft(name, "*") // package.Type
	return name
}

// StartAll iterates over a sequence of starters, starting each in
// turn, returning the first non-nil error, or nil if all starters
// returned nil.
func StartAll(ctx context.Context, starters []Starter) error {
	for _, s := range starters {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}
