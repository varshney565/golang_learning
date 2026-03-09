// Package main demonstrates the time package in Go.
// Topics: Timer, Ticker, stop leaks, time zones, formatting, monotonic clock.
// Timer/Ticker leaks are a very common interview topic.
package main

import (
	"fmt"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: time.Timer — Fire Once After a Duration
// -----------------------------------------------------------------------
// time.Timer fires ONCE after a duration.
// time.After(d) is a convenience wrapper but has a LEAK risk.
//
// LEAK RISK: time.After creates a Timer that can't be garbage collected
// until it fires. In a hot loop, this leaks memory.

func timerBasics() {
	fmt.Println("Timer basics:")

	// time.NewTimer — create and control manually
	timer := time.NewTimer(50 * time.Millisecond)
	fmt.Println("  waiting for timer...")
	<-timer.C // blocks until timer fires
	fmt.Println("  timer fired!")

	// Reset — reuse the same timer (more efficient)
	timer.Reset(30 * time.Millisecond)
	<-timer.C
	fmt.Println("  timer fired again (after reset)")
}

// -----------------------------------------------------------------------
// SECTION 2: Stopping a Timer Correctly
// -----------------------------------------------------------------------
// timer.Stop() returns false if the timer already fired.
// If Stop returns false, the channel may already have a value — drain it
// to avoid a stale receive on the next select.

func stopTimerCorrectly() {
	fmt.Println("\nStopping a timer correctly:")

	timer := time.NewTimer(100 * time.Millisecond)

	// Simulate wanting to cancel the timer early
	go func() {
		time.Sleep(20 * time.Millisecond)

		// CORRECT way to stop a timer and drain the channel
		if !timer.Stop() {
			// Timer already fired — drain the channel to prevent stale read
			select {
			case <-timer.C:
			default:
			}
		}
		fmt.Println("  timer stopped cleanly")
	}()

	select {
	case <-timer.C:
		fmt.Println("  timer fired (shouldn't happen if stopped in time)")
	case <-time.After(200 * time.Millisecond):
		fmt.Println("  confirmed: timer was stopped, didn't fire")
	}
}

// -----------------------------------------------------------------------
// SECTION 3: time.Ticker — Fire Repeatedly
// -----------------------------------------------------------------------
// time.Ticker sends on its channel at regular intervals.
// MUST call ticker.Stop() when done — otherwise goroutine leaks!

func tickerBasics() {
	fmt.Println("\nTicker basics:")

	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop() // ALWAYS stop the ticker when done

	count := 0
	for tick := range ticker.C {
		count++
		fmt.Printf("  tick %d at %s\n", count, tick.Format("15:04:05.000"))
		if count == 3 {
			break // stop after 3 ticks
		}
	}
	// After break, ticker.Stop() is called via defer
}

// -----------------------------------------------------------------------
// SECTION 4: time.After — The Leak Trap
// -----------------------------------------------------------------------
// time.After(d) returns <-chan Time — convenient but leaks if used in a loop.
// Each call allocates a new timer that can't be GC'd until it fires.

func timeAfterLeak() {
	fmt.Println("\ntime.After leak trap:")

	// BAD — in a loop: each iteration creates a new timer that lives for 1s
	// Even if select hits `default` before 1s, the timer keeps existing.
	//
	// for {
	//     select {
	//     case v := <-ch:
	//         process(v)
	//     case <-time.After(1 * time.Second): // NEW timer each iteration = LEAK
	//         log.Println("timeout")
	//     }
	// }

	// GOOD — reuse a single timer
	fmt.Println("  BAD:  time.After in a loop — creates new timer each iteration")
	fmt.Println("  GOOD: time.NewTimer outside loop, Reset inside")

	// Example of the correct pattern
	timeout := time.NewTimer(500 * time.Millisecond)
	defer timeout.Stop()

	ch := make(chan int, 1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- 42
	}()

	select {
	case v := <-ch:
		if !timeout.Stop() { // stop and drain before reuse
			<-timeout.C
		}
		fmt.Printf("  got value: %d\n", v)
	case <-timeout.C:
		fmt.Println("  timed out")
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Periodic Task with Ticker + Context
// -----------------------------------------------------------------------
// Production pattern: run a task every N seconds, stop on context cancel.

func periodicTask(done <-chan struct{}, interval time.Duration, task func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			task()
		case <-done:
			fmt.Println("  periodic task: stopped")
			return
		}
	}
}

func periodicTaskDemo() {
	fmt.Println("\nPeriodic task with done channel:")

	done := make(chan struct{})
	count := 0

	go periodicTask(done, 30*time.Millisecond, func() {
		count++
		fmt.Printf("  task ran (count=%d)\n", count)
	})

	time.Sleep(110 * time.Millisecond) // let it run ~3 times
	close(done)
	time.Sleep(20 * time.Millisecond)
}

// -----------------------------------------------------------------------
// SECTION 6: Time Formatting and Parsing
// -----------------------------------------------------------------------
// Go uses a REFERENCE TIME instead of format codes like strftime.
// The reference time is: Mon Jan 2 15:04:05 MST 2006
// (or: 01/02 03:04:05PM '06 -0700)
// Mnemonic: 1 2 3 4 5 6 7 (month day hour min sec year timezone)

func timeFormatting() {
	fmt.Println("\nTime formatting:")

	now := time.Now()

	// Format using reference time components
	fmt.Printf("  RFC3339:   %s\n", now.Format(time.RFC3339))
	fmt.Printf("  custom:    %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("  date only: %s\n", now.Format("Jan 2, 2006"))
	fmt.Printf("  time only: %s\n", now.Format("3:04 PM"))

	// Parse a time string
	t, err := time.Parse("2006-01-02", "2025-12-31")
	if err != nil {
		fmt.Printf("  parse error: %v\n", err)
	} else {
		fmt.Printf("  parsed: %v\n", t)
	}

	// Parse with location
	loc, _ := time.LoadLocation("America/New_York")
	t2, _ := time.ParseInLocation("2006-01-02 15:04:05", "2025-06-15 10:00:00", loc)
	fmt.Printf("  with timezone: %v\n", t2)
}

// -----------------------------------------------------------------------
// SECTION 7: Time Arithmetic
// -----------------------------------------------------------------------

func timeArithmetic() {
	fmt.Println("\nTime arithmetic:")

	now := time.Now()

	// Add/subtract durations
	tomorrow := now.Add(24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)
	fmt.Printf("  tomorrow:  %s\n", tomorrow.Format("2006-01-02"))
	fmt.Printf("  last week: %s\n", lastWeek.Format("2006-01-02"))

	// Duration between two times
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	elapsed := time.Since(start) // shorthand for time.Now().Sub(start)
	fmt.Printf("  elapsed:   %v\n", elapsed.Round(time.Millisecond))

	// Compare times
	t1 := now
	t2 := now.Add(time.Hour)
	fmt.Printf("  t1.Before(t2): %v\n", t1.Before(t2))
	fmt.Printf("  t2.After(t1):  %v\n", t2.After(t1))

	// Truncate and Round
	t := time.Date(2025, 1, 15, 10, 37, 45, 0, time.UTC)
	fmt.Printf("  truncate to hour: %v\n", t.Truncate(time.Hour))
	fmt.Printf("  round to hour:    %v\n", t.Round(time.Hour))
}

// -----------------------------------------------------------------------
// SECTION 8: Monotonic Clock
// -----------------------------------------------------------------------
// time.Now() returns a time with BOTH wall clock AND monotonic clock reading.
// The monotonic clock is used for measuring elapsed time accurately
// (not affected by clock adjustments like NTP sync or DST changes).
//
// When you marshal/unmarshal or compare with == you lose the monotonic reading.
// Use t.Equal(t2) instead of t == t2 for wall clock comparison.

func monotonicClock() {
	fmt.Println("\nMonotonic clock:")

	t1 := time.Now()
	time.Sleep(10 * time.Millisecond)
	t2 := time.Now()

	// Sub uses monotonic readings → accurate
	fmt.Printf("  elapsed (monotonic): %v\n", t2.Sub(t1).Round(time.Millisecond))

	// Strip monotonic reading with Round(0) or UTC()
	t3 := t1.Round(0) // strips monotonic
	fmt.Printf("  has monotonic: %v\n", t1.String())
	fmt.Printf("  stripped:      %v\n", t3.String())
}

func main() {
	timerBasics()
	stopTimerCorrectly()
	tickerBasics()
	timeAfterLeak()
	periodicTaskDemo()
	timeFormatting()
	timeArithmetic()
	monotonicClock()
}
