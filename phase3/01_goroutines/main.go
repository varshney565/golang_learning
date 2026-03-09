// Package main demonstrates goroutines in Go.
// Topics: launching goroutines, goroutine lifecycle, WaitGroup, data races.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Launching Goroutines
// -----------------------------------------------------------------------
// A goroutine is a lightweight concurrent function, managed by the Go runtime.
// Start one with the `go` keyword before a function call.
//
// Goroutines are multiplexed onto OS threads by the Go scheduler.
// You can have millions of goroutines — each starts with ~2KB stack
// (vs ~1MB for OS threads).

func sayHello(name string) {
	fmt.Printf("  Hello from %s!\n", name)
}

func basicGoroutines() {
	fmt.Println("Goroutines — basic launch:")

	// Sequential call
	sayHello("main (sequential)")

	// Concurrent calls — order is non-deterministic
	go sayHello("goroutine 1")
	go sayHello("goroutine 2")
	go sayHello("goroutine 3")

	// Problem: main() might exit before goroutines run!
	// We need to wait for them. Use time.Sleep as a crude demo only.
	time.Sleep(10 * time.Millisecond)
	// In real code, use sync.WaitGroup or channels (shown below)
}

// -----------------------------------------------------------------------
// SECTION 2: sync.WaitGroup — Proper Goroutine Waiting
// -----------------------------------------------------------------------
// WaitGroup lets the main goroutine wait for a collection of goroutines.
//   wg.Add(n)  → say "n more goroutines are running"
//   wg.Done()  → called by a goroutine when it finishes (decrements counter)
//   wg.Wait()  → blocks until counter reaches 0

func withWaitGroup() {
	fmt.Println("\nWaitGroup — proper waiting:")

	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)  // increment before launching goroutine
		i := i     // capture loop variable (Go 1.22+ fixes this, but good habit)
		go func() {
			defer wg.Done() // always decrement when done
			time.Sleep(time.Duration(i) * time.Millisecond) // simulate work
			fmt.Printf("  worker %d done\n", i)
		}()
	}

	wg.Wait() // block until all goroutines call Done()
	fmt.Println("  all workers finished")
}

// -----------------------------------------------------------------------
// SECTION 3: Goroutine Leak — A Common Bug
// -----------------------------------------------------------------------
// A goroutine leak occurs when a goroutine is started but never terminates.
// Leaks accumulate memory and CPU over time.
// Common causes: blocking channel ops with no sender/receiver, infinite loops.

func leakyGoroutine() chan int {
	ch := make(chan int)
	go func() {
		// This goroutine blocks forever waiting on ch — it's a LEAK
		// if nothing ever sends to ch
		v := <-ch
		fmt.Println(v)
	}()
	return ch
	// ch goes out of scope in caller without sending — goroutine leaks
}

// Fixed version: use a done channel to signal termination
func nonLeakyGoroutine(done <-chan struct{}) {
	ch := make(chan int)
	go func() {
		select {
		case v := <-ch:
			fmt.Println(v)
		case <-done: // exit when signaled
			fmt.Println("  goroutine stopped cleanly")
			return
		}
	}()
}

func goroutineLeak() {
	fmt.Println("\nGoroutine lifecycle:")
	fmt.Printf("  goroutines before: %d\n", runtime.NumGoroutine())

	done := make(chan struct{})
	nonLeakyGoroutine(done)

	fmt.Printf("  goroutines during: %d\n", runtime.NumGoroutine())

	close(done) // signal goroutine to stop
	time.Sleep(time.Millisecond)
	fmt.Printf("  goroutines after:  %d\n", runtime.NumGoroutine())
}

// -----------------------------------------------------------------------
// SECTION 4: Data Race — The Danger of Shared State
// -----------------------------------------------------------------------
// A data race occurs when two goroutines access the same variable
// concurrently and at least one is writing, WITHOUT synchronization.
// Results are unpredictable — undefined behavior in Go.
//
// Detect with: go run -race main.go
//              go test -race ./...

var sharedCounter int // shared state — dangerous without sync!

func dataRaceExample() {
	fmt.Println("\nData race example (UNSAFE — for illustration):")

	var wg sync.WaitGroup
	// Run the race detector: go run -race main.go
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sharedCounter++ // RACE: read-modify-write is not atomic
		}()
	}
	wg.Wait()
	fmt.Printf("  counter (expected 100, got %d — race condition!)\n", sharedCounter)
}

// -----------------------------------------------------------------------
// SECTION 5: Fixed with Mutex
// -----------------------------------------------------------------------

func fixedWithMutex() {
	fmt.Println("\nFixed with Mutex (SAFE):")

	var (
		mu      sync.Mutex
		counter int
		wg      sync.WaitGroup
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()   // acquire exclusive lock
			counter++   // safe: only one goroutine runs this at a time
			mu.Unlock() // release lock
		}()
	}
	wg.Wait()
	fmt.Printf("  counter (expected 100, got %d — correct!)\n", counter)
}

// -----------------------------------------------------------------------
// SECTION 6: Fixed with Atomic Operations
// -----------------------------------------------------------------------
// For simple counters, sync/atomic is faster than a mutex.
// Atomics guarantee read-modify-write as a single uninterruptible operation.

func fixedWithAtomic() {
	fmt.Println("\nFixed with atomic (SAFE, faster for simple ops):")
	// See phase3/04_sync/main.go for the full sync/atomic demo
	fmt.Println("  (see phase3/04_sync for atomic examples)")
}

// -----------------------------------------------------------------------
// SECTION 7: GOMAXPROCS — Parallelism Control
// -----------------------------------------------------------------------
// GOMAXPROCS controls how many OS threads run Go code simultaneously.
// Default is the number of CPU cores (Go 1.5+).
// Can be set with runtime.GOMAXPROCS(n) or GOMAXPROCS env var.

func gomaxprocsDemo() {
	fmt.Println("\nGOMAXPROCS:")
	fmt.Printf("  CPUs available:    %d\n", runtime.NumCPU())
	fmt.Printf("  GOMAXPROCS (now):  %d\n", runtime.GOMAXPROCS(0)) // 0 = query without changing
	fmt.Printf("  goroutines now:    %d\n", runtime.NumGoroutine())
}

func main() {
	basicGoroutines()
	withWaitGroup()
	goroutineLeak()
	dataRaceExample()
	fixedWithMutex()
	fixedWithAtomic()
	gomaxprocsDemo()
}
