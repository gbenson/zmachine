// Package testutil provides utilities for testing zmachine modules.
package testutil

import (
	"context"
	"fmt"
	"math"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine"
	. "gbenson.net/go/zmachine/core"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
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

const (
	DefaultAbsoluteTolerance = 1e-12
	DefaultRelativeTolerance = 1e-6
)

// NearlyEqual returns a [cmp.Comparison] that succeeds if x ≈ y.
func NearlyEqual(x, y float64) cmp.Comparison {
	return func() cmp.Result {
		if x == y {
			return cmp.ResultSuccess
		}

		delta := math.Abs(y - x)
		if delta < DefaultAbsoluteTolerance {
			return cmp.ResultSuccess
		}

		if x != 0 && delta/math.Abs(x) < DefaultRelativeTolerance {
			return cmp.ResultSuccess
		}

		return cmp.ResultFailure(fmt.Sprintf("%v !≈ %v (delta = %v)", x, y, delta))
	}
}
