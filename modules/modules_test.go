package modules

import "gbenson.net/go/zmachine/machine"

// testContext returns its receiver's context after associating
// a [logger.Logger] and a semi-configured [Machine] with it.
var testContext = machine.TestContext
