// Package testutil implements various utilities for testing the clocks.
package testutil

import (
	"time"
)

// Returns true (with the time) if the channel is readable.
// It will wait up to `timeout` before giving up and returning false.
func TryRead(ch <-chan time.Time, timeout time.Duration) (time.Time, bool) {
	var fireTime time.Time
	select {
	case fireTime = <-ch:
	case <-time.After(timeout):
	}
	return fireTime, !fireTime.IsZero()
}
