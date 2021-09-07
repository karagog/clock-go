// Package real implements a real clock, which delegates to the "time" package
// methods to access the time.
package real

import (
	"time"

	"github.com/karagog/clock-go"
)

// A real clock which calls `time.Now()` to get the time.
type Clock struct{}

func (*Clock) Now() time.Time { return time.Now() }
func (*Clock) NewTimer(d time.Duration) clock.Timer {
	return &realTimer{
		timer: time.NewTimer(d),
	}
}

// This implementation of the timer interface is backed by a real timer.
type realTimer struct {
	timer *time.Timer
}

func (t *realTimer) Reset(d time.Duration) bool {
	return t.timer.Reset(d)
}

func (t *realTimer) Stop() bool {
	return t.timer.Stop()
}

func (t *realTimer) C() <-chan time.Time {
	return t.timer.C
}
