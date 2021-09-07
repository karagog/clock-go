package simulated

import (
	"context"
	"testing"
	"time"

	"github.com/karagog/clock-go/testutil"
)

func TestSimulatedClockInitialTime(t *testing.T) {
	start := time.Now()
	c := NewClock(start)
	time.Sleep(time.Millisecond)
	if n := c.Now(); n != start {
		t.Fatalf("Got %v, want %v", n, start)
	}
}

func TestSimulatedClockAdvance(t *testing.T) {
	start := time.Now()
	c := NewClock(start)
	d := time.Second
	c.Advance(d)
	exp := start.Add(d)
	if n := c.Now(); n != exp {
		t.Fatalf("Got %v, want %v", n, exp)
	}
}

func TestSimulatedClockNegativeAdvanceDisallowed(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			return // expected
		}
		t.Fatal("Call did not panic")
	}()
	NewClock(time.Now()).Advance(-time.Nanosecond)
}

func TestSimulatedClockTimer(t *testing.T) {
	timerDuration := time.Millisecond // something short to keep the test quick
	for _, tc := range []struct {
		name     string
		step     time.Duration
		expFired bool
	}{
		{
			name:     "fire",
			step:     timerDuration,
			expFired: true,
		},
		{
			name:     "not fire",
			expFired: false,
		},
		{
			name:     "not fire upper limit",
			step:     timerDuration - time.Nanosecond,
			expFired: false,
		},
		{
			name:     "no double fire",
			step:     time.Duration(3) * timerDuration,
			expFired: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			c := NewClock(start)
			tmr := c.NewTimer(timerDuration)

			c.Advance(tc.step)

			firedTime, fired := testutil.TryRead(tmr.C(), 10*timerDuration)
			if fired {
				// Regardless of how far the timer was advanced, it should always report the correct deadline
				// when it was supposed to fire.
				if want := start.Add(timerDuration); firedTime != want {
					t.Errorf("Fired timestamp was %v (%v after start), want %v", firedTime, firedTime.Sub(start), want)
				}
				if !tc.expFired {
					t.Fatalf("Timer fired after advancing %v simulated time, wanted it to not fire.", tc.step)
				}
			} else {
				if tc.expFired {
					t.Fatalf("Timer did not fire after advancing %v simulated time, wanted it to fire.", tc.step)
				}
			}
		})
	}
}

func TestSimulatedClockTimerStopReset(t *testing.T) {
	start := time.Now()
	c := NewClock(start)
	d := time.Millisecond
	tmr := c.NewTimer(d)

	step := time.Microsecond
	c.Advance(step)
	if wasActive := tmr.Stop(); !wasActive {
		t.Fatalf("Got active %v, want true", wasActive)
	}

	// The timer should not have fired.
	if _, fired := testutil.TryRead(tmr.C(), d); fired {
		t.Fatalf("Timer fired, wanted it not to fire")
	}

	// The timer should still not fire, even if the time is advanced beyond the original deadline.
	c.Advance(d)
	if _, fired := testutil.TryRead(tmr.C(), d); fired {
		t.Fatalf("Timer fired after it has already been stopped, wanted it not to fire")
	}

	// Reset the timer for the next part of the test.
	if wasActive := tmr.Reset(d); wasActive {
		t.Fatal("Timer said it was active, want not active")
	}

	// Now advance the time so the timer fires and check the result of Stop().
	c.Advance(d)
	if wasActive := tmr.Stop(); wasActive {
		t.Fatal("Timer said it was active, want not active")
	}
	if _, fired := testutil.TryRead(tmr.C(), time.Second); !fired {
		t.Fatalf("Timer not fired, want fired")
	}

	// Resetting a timer with 0 duration should fire immediately.
	if wasActive := tmr.Reset(0); wasActive {
		t.Fatalf("Got timer active, want inactive")
	}
	if _, fired := testutil.TryRead(tmr.C(), time.Second); !fired {
		t.Fatalf("Timer not fired, want fired")
	}
}

func TestSimulatedClockMultipleTimers(t *testing.T) {
	start := time.Now()
	c := NewClock(start)
	d := time.Millisecond

	// There is no way to guarantee the order in which they fired, due to the nature of
	// select statements with buffered channels, but we can verify they all fired.
	t1 := c.NewTimer(d)
	t2 := c.NewTimer(2 * d)

	c.Advance(time.Second)

	for cnt := 0; cnt < 2; cnt++ {
		var got, want time.Time
		select {
		case got = <-t1.C():
			want = start.Add(d)
		case got = <-t2.C():
			want = start.Add(2 * d)
		default:
			t.Fatal("Not all timers fired")
		}
		if got != want {
			t.Errorf("Got time %v, want %v", got, want)
		}
	}
}

func TestSimulatedClockMultithreaded(t *testing.T) {
	// These flags can help debug tricky multithreading errors.
	//flag.Set("alsologtostderr", "true")
	//flag.Set("v", "4")

	start := time.Date(2021, 3, 22, 0, 0, 0, 0, time.UTC)
	c := NewClock(start)
	d := time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	count := 5                    // trigger the timer this many times
	fired := make(chan time.Time) // every time the timer fires it will be written here
	tmr := c.NewTimer(d)          // create outside the goroutine to prevent a data race with the test
	go func() {
		for i := 0; i < count; i++ {
			select {
			case tm := <-tmr.C():
				tmr.Reset(d)
				fired <- tm
			case <-ctx.Done():
			}
		}
	}()

	for i := 0; i < count; i++ {
		c.Advance(d)
		exp := start.Add(time.Duration(i+1) * d)
		select {
		case tm := <-fired:
			if tm != exp {
				t.Fatalf("Got %v on iteration %v, want %v", tm, i, exp)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("Background timer did not fire on iteration %d", i)
		}
	}
}
