// Package main demonstrates deadlock scenarios in Go — and how to fix them.
// Deadlocks are a top interview topic. Know how to spot and fix each type.
//
// NOTE: Actual deadlocks will crash the program with "all goroutines are asleep".
//       The deadlock examples are shown as commented code with explanations.
//       Safe alternatives are demonstrated as runnable code.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Classic Deadlock — Unbuffered Channel, Same Goroutine
// -----------------------------------------------------------------------
// Deadlock: a goroutine blocks waiting for itself.

func classicDeadlock() {
	fmt.Println("Classic channel deadlock (explained, not run):")
	fmt.Println(`
  DEADLOCK CODE:
    ch := make(chan int)
    ch <- 1      // ← blocks waiting for a receiver
    <-ch         // ← never reached
    // fatal error: all goroutines are asleep - deadlock!

  WHY: Unbuffered send blocks until someone receives.
       Same goroutine can't both send and receive.

  FIX 1: Use a buffered channel
    ch := make(chan int, 1)
    ch <- 1   // doesn't block (buffer has space)
    <-ch      // works

  FIX 2: Send from a separate goroutine
    ch := make(chan int)
    go func() { ch <- 1 }()
    <-ch
`)

	// Fixed version
	ch := make(chan int, 1)
	ch <- 1
	v := <-ch
	fmt.Printf("  Fixed: received %d\n", v)
}

// -----------------------------------------------------------------------
// SECTION 2: Mutex Deadlock — Lock Called Twice
// -----------------------------------------------------------------------

func mutexDeadlock() {
	fmt.Println("\nMutex deadlock (explained, not run):")
	fmt.Println(`
  DEADLOCK CODE:
    var mu sync.Mutex
    mu.Lock()
    mu.Lock()   // ← blocks forever, same goroutine
    mu.Unlock()

  WHY: sync.Mutex is not reentrant — same goroutine can't lock it twice.

  FIX 1: Don't call Lock twice in same goroutine
  FIX 2: Use a separate mutex for the inner critical section
  FIX 3: Refactor to remove the nested lock need
  FIX 4: Use sync.RWMutex with RLock in safe read paths
`)

	// Safe demonstration
	var mu sync.Mutex
	mu.Lock()
	fmt.Println("  first lock acquired")
	mu.Unlock()
	mu.Lock()
	fmt.Println("  second lock acquired (after unlock)")
	mu.Unlock()
}

// -----------------------------------------------------------------------
// SECTION 3: Mutex Deadlock — A Locks B, B Locks A (Circular Wait)
// -----------------------------------------------------------------------

type Account struct {
	mu      sync.Mutex
	balance float64
	name    string
}

// UNSAFE: deadlock if called concurrently as transfer(A,B) and transfer(B,A)
func transferUnsafe(from, to *Account, amount float64) {
	from.mu.Lock()
	defer from.mu.Unlock()
	to.mu.Lock() // DEADLOCK: goroutine 1 holds A, waits for B
	defer to.mu.Unlock() //         goroutine 2 holds B, waits for A
	from.balance -= amount
	to.balance += amount
}

// SAFE: always acquire locks in the same order (by pointer address or name)
func transferSafe(from, to *Account, amount float64) {
	// Lock in consistent order — prevents circular wait
	first, second := from, to
	if from.name > to.name { // arbitrary but consistent ordering
		first, second = to, from
	}
	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()

	from.balance -= amount
	to.balance += amount
}

func circularDeadlock() {
	fmt.Println("\nCircular mutex deadlock (explained):")
	fmt.Println(`
  DEADLOCK SCENARIO:
    // goroutine 1:
    transfer(accountA, accountB, 100)  // locks A, then tries to lock B
    // goroutine 2 (concurrent):
    transfer(accountB, accountA, 50)   // locks B, then tries to lock A
    // Both goroutines wait for each other → deadlock!

  FIX: Always acquire multiple locks in the SAME ORDER.
       E.g., always lock lower-address pointer first, or sort by ID/name.
`)

	a := &Account{name: "Alice", balance: 1000}
	b := &Account{name: "Bob", balance: 500}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			transferSafe(a, b, 10)
		}()
		go func() {
			defer wg.Done()
			transferSafe(b, a, 5)
		}()
	}
	wg.Wait()
	fmt.Printf("  Alice: %.0f, Bob: %.0f (no deadlock)\n", a.balance, b.balance)
}

// -----------------------------------------------------------------------
// SECTION 4: Channel Deadlock — Forgetting to Close
// -----------------------------------------------------------------------

func forgotToClose() {
	fmt.Println("\nForgetting to close channel (explained):")
	fmt.Println(`
  DEADLOCK CODE:
    ch := make(chan int)
    go func() {
        ch <- 1
        ch <- 2
        // forgot: close(ch)
    }()
    for v := range ch {  // ← blocks forever after receiving 2
        fmt.Println(v)
    }

  WHY: range on channel blocks until channel is CLOSED.
       If sender never closes, consumer waits forever.

  FIX: Always close the channel when done sending.
`)

	// Fixed version
	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 2
		close(ch) // FIX: signal "no more values"
	}()
	for v := range ch {
		fmt.Printf("  received %d\n", v)
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Goroutine Leak — Blocked Forever on Channel
// -----------------------------------------------------------------------
// Not a classic deadlock (program doesn't crash) but goroutines never exit.

func goroutineLeakDemo() {
	fmt.Println("\nGoroutine leak (not a crash, but a resource leak):")

	fmt.Printf("  goroutines before: %d\n", goroutineCount())

	// LEAK: goroutine blocks forever waiting on ch that nobody writes to
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		go func() {
			<-ch // blocks forever
		}()
	}
	time.Sleep(10 * time.Millisecond) // let goroutines start
	fmt.Printf("  goroutines during leak: %d\n", goroutineCount())

	// FIX: use a done channel or context to stop them
	done := make(chan struct{})
	ch2 := make(chan int)
	for i := 0; i < 3; i++ {
		go func() {
			select {
			case v := <-ch2:
				_ = v
			case <-done: // exit when told
				return
			}
		}()
	}
	time.Sleep(10 * time.Millisecond)
	close(done) // stop all goroutines
	time.Sleep(10 * time.Millisecond)
	// Note: the leaked goroutines from ch above are still running
	fmt.Printf("  goroutines after fix: %d (leaked ones still exist)\n", goroutineCount())

	// Close the leak channel to finally unblock the leaked goroutines
	close(ch)
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("  goroutines after closing leak: %d\n", goroutineCount())
}

func goroutineCount() int {
	return runtime.NumGoroutine()
}

// -----------------------------------------------------------------------
// SECTION 6: Select Deadlock — All Channels Block
// -----------------------------------------------------------------------

func selectDeadlock() {
	fmt.Println("\nSelect with all channels blocking (explained):")
	fmt.Println(`
  DEADLOCK CODE:
    ch1 := make(chan int)
    ch2 := make(chan int)
    select {
    case v := <-ch1:  // ch1 is empty, blocks
        _ = v
    case v := <-ch2:  // ch2 is empty, blocks
        _ = v
    }
    // fatal error: all goroutines are asleep - deadlock!

  WHY: select blocks when ALL cases block and there's no default.

  FIX: Add a default case for non-blocking select.
       OR ensure at least one channel will be ready.
`)

	// Fixed with default
	ch1 := make(chan int)
	ch2 := make(chan int)
	select {
	case v := <-ch1:
		fmt.Printf("  ch1: %d\n", v)
	case v := <-ch2:
		fmt.Printf("  ch2: %d\n", v)
	default:
		fmt.Println("  no channels ready — default executed (no deadlock)")
	}
}

// -----------------------------------------------------------------------
// SECTION 7: WaitGroup Deadlock — Add Called After Wait
// -----------------------------------------------------------------------

func waitGroupDeadlock() {
	fmt.Println("\nWaitGroup misuse (explained):")
	fmt.Println(`
  DEADLOCK CODE:
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        go func() {
            wg.Add(1)      // ← ADD inside goroutine
            defer wg.Done()
            // ... work
        }()
    }
    wg.Wait()   // may return before Add is called!
    // Sometimes works (race), sometimes deadlock (Wait finishes early)

  WHY: If wg.Wait() is called before the goroutines call wg.Add(1),
       Wait returns immediately because counter is 0.
       Then goroutines try to call Done with counter at 0 → panic.

  FIX: Always call wg.Add BEFORE launching the goroutine.
`)

	// Fixed version
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1) // ADD before go func
		i := i
		go func() {
			defer wg.Done()
			fmt.Printf("  worker %d done\n", i)
		}()
	}
	wg.Wait()
	fmt.Println("  all workers done (no deadlock)")
}

// -----------------------------------------------------------------------
// SECTION 8: Detecting Deadlocks
// -----------------------------------------------------------------------

func detectionMethods() {
	fmt.Println("\nDeadlock detection methods:")
	methods := []struct{ method, desc string }{
		{"Go runtime", "Automatically detects when ALL goroutines are stuck → prints stack + exits"},
		{"go test -race", "Detects data races that can lead to deadlocks"},
		{"pprof goroutine profile", "Shows all goroutine stacks — find blocked ones"},
		{"context.WithTimeout", "Don't block forever — use timeouts everywhere"},
		{"go vet", "Catches some misuse patterns (copies of mutexes, etc.)"},
	}
	for _, m := range methods {
		fmt.Printf("  %-30s → %s\n", m.method, m.desc)
	}
}

func main() {
	classicDeadlock()
	mutexDeadlock()
	circularDeadlock()
	forgotToClose()
	selectDeadlock()
	waitGroupDeadlock()
	detectionMethods()
}
