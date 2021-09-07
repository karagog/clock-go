package real

import (
	"testing"
	"time"

	"github.com/karagog/clock-go/testutil"
)

func TestRealClock(t *testing.T) {
	c := Clock{}
	n := c.Now()
	now := time.Now()
	diff := now.Sub(n)
	tolerance := time.Millisecond
	if diff > tolerance {
		t.Fatalf("Got time %v, want %v to within %v tolerance. Diff was %v.",
			n, now, tolerance, diff)
	}
}

// No need to test this thoroughly, since it's backed by Golang's real timer.
// We'll just test it enough to show that it doesn't crash.
func TestRealClockTimer(t *testing.T) {
	c := Clock{}
	d := time.Millisecond
	tmr := c.NewTimer(d)
	if _, fired := testutil.TryRead(tmr.C(), time.Second); !fired {
		t.Fatalf("Timer did not fire after %v, wanted it to fire", time.Second)
	}
	if tmr.Stop() {
		<-tmr.C()
	}
	tmr.Reset(time.Second)
}
