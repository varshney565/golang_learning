// Package main demonstrates sync.Cond — condition variables.
// sync.Cond is asked in interviews but misunderstood by many.
// Topics: Wait/Signal/Broadcast, producer-consumer, bounded buffer.
package main

import (
	"fmt"
	"sync"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: What is sync.Cond?
// -----------------------------------------------------------------------
// sync.Cond lets goroutines WAIT until a condition becomes true,
// without busy-polling.
//
// Three operations:
//   cond.Wait()      — atomically releases the lock AND suspends the goroutine
//                      when woken, re-acquires the lock before returning
//   cond.Signal()    — wake ONE waiting goroutine
//   cond.Broadcast() — wake ALL waiting goroutines
//
// Rule: ALWAYS call Wait() inside a loop that re-checks the condition.
//       A goroutine can wake up spuriously (without Signal/Broadcast).
//
//   mu.Lock()
//   for !condition {   ← loop, not if
//       cond.Wait()
//   }
//   // ... use shared state
//   mu.Unlock()

// -----------------------------------------------------------------------
// SECTION 2: Basic Signal — One Producer, One Consumer
// -----------------------------------------------------------------------

func basicSignal() {
	fmt.Println("Basic Signal (one producer, one consumer):")

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false

	// Consumer — waits until producer signals
	go func() {
		mu.Lock()
		for !ready { // loop: re-check condition after waking
			fmt.Println("  consumer: waiting for data...")
			cond.Wait() // releases mu, suspends; re-acquires mu on wake
		}
		fmt.Println("  consumer: got the signal, processing!")
		mu.Unlock()
	}()

	// Producer — does work, then signals
	time.Sleep(100 * time.Millisecond) // simulate work
	mu.Lock()
	ready = true
	fmt.Println("  producer: data ready, signaling...")
	cond.Signal() // wake one waiting goroutine
	mu.Unlock()

	time.Sleep(50 * time.Millisecond) // let consumer finish
}

// -----------------------------------------------------------------------
// SECTION 3: Broadcast — Many Waiters
// -----------------------------------------------------------------------
// Use Broadcast when ALL waiting goroutines should wake up.
// Example: config reload, cache invalidation, gate open.

func broadcastExample() {
	fmt.Println("\nBroadcast (gate opens for all goroutines):")

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	gateOpen := false

	var wg sync.WaitGroup
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			mu.Lock()
			for !gateOpen {
				cond.Wait()
			}
			fmt.Printf("  goroutine %d: gate is open, proceeding!\n", i)
			mu.Unlock()
		}()
	}

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	gateOpen = true
	fmt.Println("  main: opening gate for all...")
	cond.Broadcast() // wake ALL waiters at once
	mu.Unlock()

	wg.Wait()
}

// -----------------------------------------------------------------------
// SECTION 4: Bounded Buffer (Classic sync.Cond Use Case)
// -----------------------------------------------------------------------
// A buffer with a max capacity.
// Producers block when full; consumers block when empty.
// This is more efficient than a buffered channel when you need
// fine-grained control over the buffer behavior.

type BoundedBuffer struct {
	mu       sync.Mutex
	notFull  *sync.Cond // signal: "there's now space in the buffer"
	notEmpty *sync.Cond // signal: "there's now data in the buffer"
	buf      []int
	cap      int
}

func NewBoundedBuffer(capacity int) *BoundedBuffer {
	bb := &BoundedBuffer{cap: capacity}
	bb.notFull = sync.NewCond(&bb.mu)
	bb.notEmpty = sync.NewCond(&bb.mu)
	return bb
}

func (bb *BoundedBuffer) Put(item int) {
	bb.mu.Lock()
	for len(bb.buf) == bb.cap { // buffer is full — wait
		bb.notFull.Wait()
	}
	bb.buf = append(bb.buf, item)
	bb.notEmpty.Signal() // tell consumers there's data
	bb.mu.Unlock()
}

func (bb *BoundedBuffer) Get() int {
	bb.mu.Lock()
	for len(bb.buf) == 0 { // buffer is empty — wait
		bb.notEmpty.Wait()
	}
	item := bb.buf[0]
	bb.buf = bb.buf[1:]
	bb.notFull.Signal() // tell producers there's space
	bb.mu.Unlock()
	return item
}

func boundedBuffer() {
	fmt.Println("\nBounded Buffer (capacity=3):")

	buf := NewBoundedBuffer(3)
	var wg sync.WaitGroup

	// Producer — tries to put 6 items into capacity-3 buffer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 6; i++ {
			buf.Put(i)
			fmt.Printf("  put: %d\n", i)
		}
	}()

	// Consumer — slowly consumes
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 6; i++ {
			time.Sleep(30 * time.Millisecond) // simulate slow consumer
			v := buf.Get()
			fmt.Printf("  got: %d\n", v)
		}
	}()

	wg.Wait()
}

// -----------------------------------------------------------------------
// SECTION 5: sync.Cond vs Channel — When to Use Each
// -----------------------------------------------------------------------
//
// Use CHANNEL when:
//   - Transferring data between goroutines
//   - Simple signaling (done, cancel)
//   - Fan-out / pipeline patterns
//
// Use sync.COND when:
//   - Multiple goroutines wait on the SAME shared state
//   - You need Broadcast (channels can't broadcast efficiently)
//   - You need to check a complex condition atomically with a lock
//   - Implementing classic concurrent data structures (bounded buffer, etc.)

func comparison() {
	fmt.Println("\nsync.Cond vs Channel:")
	fmt.Println("  Channel:   transferring data, simple signal, fan-out")
	fmt.Println("  sync.Cond: multiple waiters on shared state, Broadcast, complex conditions")
}

func main() {
	basicSignal()
	broadcastExample()
	boundedBuffer()
	comparison()
}
