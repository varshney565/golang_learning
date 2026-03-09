// Package main demonstrates the select statement in Go.
// Topics: select basics, default, timeout, fan-in, priority select.
package main

import (
	"fmt"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: select Basics
// -----------------------------------------------------------------------
// select waits on multiple channel operations simultaneously.
// It's like a switch statement, but for channels.
//
//   - Blocks until ONE of the cases is ready
//   - If multiple cases are ready simultaneously → one is chosen AT RANDOM
//   - This intentional randomness avoids starvation

func selectBasics() {
	fmt.Println("select basics:")

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	ch1 <- "one"
	ch2 <- "two"

	// Both are ready — Go will randomly pick one
	select {
	case msg := <-ch1:
		fmt.Printf("  received from ch1: %s\n", msg)
	case msg := <-ch2:
		fmt.Printf("  received from ch2: %s\n", msg)
	}
}

// -----------------------------------------------------------------------
// SECTION 2: select with default — Non-Blocking
// -----------------------------------------------------------------------
// Adding a `default` case makes select non-blocking.
// If no channel is ready, default runs immediately.
// Use this to poll channels without blocking.

func nonBlockingSelect() {
	fmt.Println("\nNon-blocking select (with default):")

	ch := make(chan int, 1)

	// Nothing in ch yet — default fires
	select {
	case v := <-ch:
		fmt.Printf("  received: %d\n", v)
	default:
		fmt.Println("  no value ready, continuing")
	}

	ch <- 42

	// Now ch has a value — case fires
	select {
	case v := <-ch:
		fmt.Printf("  received: %d\n", v)
	default:
		fmt.Println("  no value ready, continuing")
	}
}

// -----------------------------------------------------------------------
// SECTION 3: Timeout Pattern
// -----------------------------------------------------------------------
// time.After(d) returns a channel that receives after duration d.
// Combine with select to implement timeouts.

func withTimeout(name string, delay time.Duration, timeout time.Duration) {
	result := make(chan string, 1)

	go func() {
		time.Sleep(delay) // simulate work
		result <- "result from " + name
	}()

	select {
	case r := <-result:
		fmt.Printf("  %s: %s\n", name, r)
	case <-time.After(timeout):
		fmt.Printf("  %s: TIMEOUT after %v\n", name, timeout)
	}
}

func timeoutPattern() {
	fmt.Println("\nTimeout pattern:")
	withTimeout("fast-service", 10*time.Millisecond, 50*time.Millisecond)  // finishes in time
	withTimeout("slow-service", 100*time.Millisecond, 50*time.Millisecond) // times out
}

// -----------------------------------------------------------------------
// SECTION 4: Fan-In (Multiplexing)
// -----------------------------------------------------------------------
// Merge multiple channels into one using select in a loop.
// Useful when you have multiple producers and one consumer.

func fanIn(ch1, ch2 <-chan string) <-chan string {
	merged := make(chan string)
	go func() {
		defer close(merged)
		// We need to track when both channels are done
		ch1Done, ch2Done := false, false
		for !ch1Done || !ch2Done {
			select {
			case v, ok := <-ch1:
				if !ok {
					ch1Done = true
					ch1 = nil // nil channel blocks forever in select — disables this case
					continue
				}
				merged <- v
			case v, ok := <-ch2:
				if !ok {
					ch2Done = true
					ch2 = nil
					continue
				}
				merged <- v
			}
		}
	}()
	return merged
}

func fanInDemo() {
	fmt.Println("\nFan-in (merging two channels):")

	makeSource := func(name string, msgs ...string) <-chan string {
		ch := make(chan string)
		go func() {
			for _, m := range msgs {
				time.Sleep(10 * time.Millisecond)
				ch <- fmt.Sprintf("[%s] %s", name, m)
			}
			close(ch)
		}()
		return ch
	}

	src1 := makeSource("A", "hello", "world")
	src2 := makeSource("B", "foo", "bar", "baz")

	for msg := range fanIn(src1, src2) {
		fmt.Printf("  %s\n", msg)
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Nil Channel in select
// -----------------------------------------------------------------------
// A receive on a nil channel BLOCKS FOREVER.
// This is useful to disable a select case dynamically.

func nilChannelInSelect() {
	fmt.Println("\nNil channel disabling select case:")

	ch1 := make(chan string, 2)
	ch2 := make(chan string, 2)

	ch1 <- "from ch1"
	ch2 <- "from ch2"

	// Process both channels, disabling each when exhausted
	for i := 0; i < 2; i++ {
		select {
		case v, ok := <-ch1:
			if !ok {
				ch1 = nil // disable this case
			} else {
				fmt.Printf("  ch1: %s\n", v)
			}
		case v, ok := <-ch2:
			if !ok {
				ch2 = nil
			} else {
				fmt.Printf("  ch2: %s\n", v)
			}
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Done Channel with select (Cancellation)
// -----------------------------------------------------------------------
// The standard pattern to cancel a long-running goroutine.

func longRunningWork(done <-chan struct{}, results chan<- int) {
	n := 0
	for {
		select {
		case <-done:
			fmt.Printf("  worker stopping after %d iterations\n", n)
			return
		default:
			// Do a unit of work
			n++
			// In real code: process something, then loop
		}
		if n >= 1000 { // safety limit for demo
			break
		}
	}
	results <- n
}

func doneWithSelect() {
	fmt.Println("\nCancellation with done channel:")

	done := make(chan struct{})
	results := make(chan int, 1)

	go longRunningWork(done, results)

	time.Sleep(time.Millisecond)
	close(done) // cancel the worker

	// Wait a bit for the worker to notice the cancellation
	time.Sleep(time.Millisecond)
}

func main() {
	selectBasics()
	nonBlockingSelect()
	timeoutPattern()
	fanInDemo()
	nilChannelInSelect()
	doneWithSelect()
}
