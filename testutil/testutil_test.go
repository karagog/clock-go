package testutil

import (
	"testing"
	"time"
)

func TestTryRead(t *testing.T) {
	ch := make(chan time.Time, 1) // allow buffering to avoid more goroutines
	tm, ok := TryRead(ch, time.Millisecond)
	if ok {
		t.Fatalf("Got channel readable (%v), want not readable", tm)
	}

	writeTime := time.Now()
	ch <- writeTime
	tm, ok = TryRead(ch, time.Millisecond)
	if !ok {
		t.Fatalf("Got channel not readable, want %q", writeTime)
	}
	if got, want := tm, writeTime; got != want {
		t.Fatalf("Got time %q, want %q", got, want)
	}
}
