// Package testutil provides utilities for testing zmachine modules.
package testutil

import (
	"context"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
	"gotest.tools/v3/assert"
)

// Tester is the subset of [testing.T] et al used by this package.
type Tester interface {
	logger.Contexter
	assert.TestingT
	Helper()
}

// TestContext returns its receiver's context after associating a
// [logger.Logger] and a semi-configured [zmachine.Machine] with it.
// The receiver should be a [testing.T] or similar.
func TestContext(t logger.Contexter) context.Context {
	return zmachine.New().WithContext(logger.TestContext(t))
}

// TestStart starts a [Starter] with a [TestContext], failing
// the test immediately if the starter returns a non-nil error.
// The receiver should be a [testing.T] or similar.
func StartForTest(t Tester, s Starter) {
	t.Helper()
	assert.NilError(t, s.Start(TestContext(t)))
}
