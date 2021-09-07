// Package simulated implements a simulated clock for unit tests.
package simulated

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/karagog/clock-go"
)

// Clock is a fake clock that you can manipulate through the accessor methods.
//
// Create with NewClock(). Advance time with Advance().
type Clock struct {
	mu           sync.RWMutex // guards all members below
	now          time.Time
	activeTimers map[*simulatedTimer]bool
}

func NewClock(now time.Time) *Clock {
	return &Clock{
		now:          now,
		activeTimers: make(map[*simulatedTimer]bool),
	}
}

func (s *Clock) Advance(d time.Duration) {
	if d < 0 {
		panic("Time can only advance in the positive direction")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	newTime := s.now.Add(d)
	glog.V(1).Infof("Advancing the time %v from %v to %v", newTime.Sub(s.now), s.now, newTime)
	s.now = newTime
	for t := range s.activeTimers {
		if t.shouldFire(s.now) {
			t.fire() // timer fired!
		}
	}
}

func (s *Clock) Now() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.now
}

func (s *Clock) NewTimer(d time.Duration) clock.Timer {
	t := &simulatedTimer{
		c:     make(chan time.Time, 1), // small buffer to avoid blocking
		clock: s,
	}
	t.Reset(d)
	return t
}

// The timer object may run on a separate goroutine from the clock object, but
// it should never be used from multiple goroutines simultaneously.
type simulatedTimer struct {
	clock    *Clock
	c        chan time.Time
	deadline time.Time // when the timer should fire
	active   bool      // true if the timer has not yet fired
}

func (t *simulatedTimer) Reset(d time.Duration) bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	wasActive := t.active
	t.active = true
	now := t.clock.now
	newDeadline := now.Add(d)
	glog.V(1).Infof("Resetting timer from deadline %v to %v", t.deadline, newDeadline)
	t.deadline = newDeadline
	if t.shouldFire(now) {
		t.fire()
		return wasActive
	}
	if !wasActive {
		glog.V(1).Infof("Timer became active")
		t.clock.activeTimers[t] = true
	}
	return wasActive
}

func (t *simulatedTimer) shouldFire(now time.Time) bool {
	return !now.Before(t.deadline)
}

func (t *simulatedTimer) fire() {
	glog.V(1).Infof("Timer fired at %v", t.deadline)
	t.c <- t.deadline
	t.stopNoLock()
}

func (t *simulatedTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	return t.stopNoLock()
}

// LOCK_REQUIRED t.clock.mu
func (t *simulatedTimer) stopNoLock() bool {
	wasActive := t.active
	t.active = false
	delete(t.clock.activeTimers, t)
	return wasActive
}

func (t *simulatedTimer) C() <-chan time.Time {
	return t.c
}
