// Package main demonstrates channels in Go.
// Topics: unbuffered, buffered, direction, range over channel, closing.
package main

import (
	"fmt"
	"sync"
)

// -----------------------------------------------------------------------
// SECTION 1: Unbuffered Channels
// -----------------------------------------------------------------------
// An unbuffered channel (capacity 0) is a synchronization point.
//
//   ch <- value  → BLOCKS until someone receives
//   <-ch         → BLOCKS until someone sends
//
// Think of it as a handshake — sender and receiver must both be ready.

func unbufferedChannels() {
	fmt.Println("Unbuffered channels:")

	ch := make(chan int) // unbuffered: make(chan int) or make(chan int, 0)

	// Launch goroutine to send — it will block until main receives
	go func() {
		fmt.Println("  goroutine: about to send 42")
		ch <- 42 // blocks here until main receives
		fmt.Println("  goroutine: send complete")
	}()

	fmt.Println("  main: about to receive")
	v := <-ch // blocks until goroutine sends
	fmt.Printf("  main: received %d\n", v)
}

// -----------------------------------------------------------------------
// SECTION 2: Buffered Channels
// -----------------------------------------------------------------------
// A buffered channel has a queue. Sends don't block until the buffer is FULL.
// Receives don't block unless the buffer is EMPTY.
//
// Use buffered channels when:
//   - You know the maximum number of items in flight
//   - You want to decouple sender/receiver timing
//   - Implementing semaphores

func bufferedChannels() {
	fmt.Println("\nBuffered channels:")

	ch := make(chan string, 3) // buffer of 3

	// These sends don't block — buffer has room
	ch <- "first"
	ch <- "second"
	ch <- "third"
	fmt.Printf("  sent 3 items, len=%d cap=%d\n", len(ch), cap(ch))

	// Receive them in order (FIFO)
	fmt.Printf("  received: %s\n", <-ch)
	fmt.Printf("  received: %s\n", <-ch)
	fmt.Printf("  received: %s\n", <-ch)

	// A 4th send without receives would block:
	// ch <- "fourth"  ← deadlock if no goroutine is receiving
}

// -----------------------------------------------------------------------
// SECTION 3: Channel Direction — Read-Only and Write-Only
// -----------------------------------------------------------------------
// You can restrict a channel to only sends or only receives in function params.
// This makes intent clear and prevents misuse at compile time.
//
//   chan<- T   → send-only (can only write to it)
//   <-chan T   → receive-only (can only read from it)

// producer only sends to ch — declared as send-only
func producer(ch chan<- int, n int) {
	for i := 0; i < n; i++ {
		ch <- i
	}
	close(ch) // important: signal no more values
}

// consumer only receives from ch — declared as receive-only
func consumer(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for v := range ch { // range on channel: receives until channel is closed
		fmt.Printf("  consumed: %d\n", v)
	}
}

func channelDirection() {
	fmt.Println("\nChannel direction:")

	ch := make(chan int, 5)
	var wg sync.WaitGroup

	wg.Add(1)
	go consumer(ch, &wg) // pass as receive-only
	producer(ch, 5)       // pass as send-only (also closes ch)

	wg.Wait()
}

// -----------------------------------------------------------------------
// SECTION 4: Closing Channels — Rules and Patterns
// -----------------------------------------------------------------------
// Rules for closing channels:
//   1. Only the SENDER should close a channel (not the receiver)
//   2. Sending to a closed channel PANICS
//   3. Receiving from a closed channel returns zero value immediately
//   4. Use "comma ok" to check if channel is closed: v, ok := <-ch
//   5. range over a channel exits when the channel is closed

func closingChannels() {
	fmt.Println("\nClosing channels:")

	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	close(ch) // close after all sends

	// "comma ok" form — ok=false means channel is closed and empty
	for {
		v, ok := <-ch
		if !ok {
			fmt.Printf("  channel closed (ok=false)\n")
			break
		}
		fmt.Printf("  received %d (ok=%v)\n", v, ok)
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Channel as Semaphore
// -----------------------------------------------------------------------
// A buffered channel of size N acts as a semaphore, limiting concurrency.
// Common pattern: limit the number of concurrent goroutines.

func withSemaphore() {
	fmt.Println("\nSemaphore pattern (limit to 2 concurrent):")

	// Semaphore: buffer = max concurrent goroutines
	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			sem <- struct{}{} // acquire: blocks if 2 slots already taken
			defer func() { <-sem }() // release when done

			fmt.Printf("  task %d running\n", i)
		}()
	}

	wg.Wait()
}

// -----------------------------------------------------------------------
// SECTION 6: Done Channel Pattern
// -----------------------------------------------------------------------
// A common pattern: use a channel to signal goroutines to stop.
// Closing a channel broadcasts to ALL receivers simultaneously.

func doneChannel() {
	fmt.Println("\nDone channel broadcast:")

	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			select {
			case <-done:
				fmt.Printf("  worker %d: received stop signal\n", i)
			}
		}()
	}

	// Close broadcasts to all 3 goroutines at once
	close(done)
	wg.Wait()
}

// -----------------------------------------------------------------------
// SECTION 7: Pipeline Pattern
// -----------------------------------------------------------------------
// Channels connect stages of a data processing pipeline.
// Each stage: receives from input, transforms, sends to output.

func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func pipeline() {
	fmt.Println("\nPipeline pattern:")

	// Chain: generate → square → print
	nums := generate(1, 2, 3, 4, 5)
	squares := square(nums)

	for v := range squares {
		fmt.Printf("  %d\n", v)
	}
}

func main() {
	unbufferedChannels()
	bufferedChannels()
	channelDirection()
	closingChannels()
	withSemaphore()
	doneChannel()
	pipeline()
}
