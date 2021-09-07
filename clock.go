// Package clock implements a clock interface.
//
// Your application code should use the clock interface to interact with time wherever possible,
// and then it can be easily unit-tested by swapping out the "real" clock with a "simulated"
// one during unit tests via dependency injection techniques.
package clock

import (
	"time"
)

// An abstract clock interface to facilitate dependency injection of time.
//
// Implementations must provide thread-safe access to interface methods, to allow
// clocks to be used in background goroutines.
type Clock interface {
	// Returns the current time.
	Now() time.Time

	// Returns a timer that will fire after the given duration from Now().
	NewTimer(time.Duration) Timer
}

// This has essentially the same interface as a time.Timer, except C() is a function
// in order to make it a pure interface that we can mock.
type Timer interface {
	Reset(time.Duration) bool
	Stop() bool
	C() <-chan time.Time
}
