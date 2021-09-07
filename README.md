# clock-go
Provides a "Clock" interface for facilitating dependency injection in time-related Go code.

## Why?

Golang's "time" package does not facilitate dependency injection by itself, because it relies on package-level functions (e.g. time.Now()). It can be very tricky to properly unit test code like that, because it may require carefully choreographed sleep statements and may end up non-deterministic/flaky. Using a dependency-injected clock allows you to use a real clock in production and a deterministic simulated clock in unit tests.

## Example

```go
package main

import (
  "github.com/karagog/clock-go"
  "github.com/karagog/clock-go/real"
  "github.com/karagog/clock-go/simulated"
)

// Given a function that needs the current time, inject a clock
// object instead of calling "time.Now()".
func DoIt(c clock.Clock) {
  fmt.Printf("The time is %v", c.Now())
}

func main() {
  // You can inject a real clock, which delegates to time.Now().
  DoIt(&real.Clock{})

  // ...or you can use a simulated clock, whose output is
  // deterministic and controlled by you.
  //
  // For example, this simulated clock is one hour in the future:
  s := simulated.NewClock(time.Now().Add(time.Hour))
  DoIt(s)

  // You can advance the time in any increments you want.
  s.Advance(time.Second)
  DoIt(s)
}
```
