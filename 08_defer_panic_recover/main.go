// Package main demonstrates defer, panic, and recover in Go.
// Topics: defer order, defer with closures, panic, recover, real-world patterns.
package main

import "fmt"

// -----------------------------------------------------------------------
// SECTION 1: defer — Basics
// -----------------------------------------------------------------------
// `defer` schedules a function call to run AFTER the surrounding function
// returns (or panics). Used for cleanup: closing files, unlocking mutexes,
// ending timers, etc.
//
// Key rules:
//   1. Deferred calls execute in LIFO order (last in, first out)
//   2. Arguments to deferred functions are evaluated IMMEDIATELY (at defer time)
//   3. Deferred functions CAN access and modify named return values

func deferBasics() {
	fmt.Println("defer basics:")
	fmt.Println("  step 1")
	defer fmt.Println("  deferred A") // runs last
	fmt.Println("  step 2")
	defer fmt.Println("  deferred B") // runs second-to-last
	fmt.Println("  step 3")
	defer fmt.Println("  deferred C") // runs first among deferred (LIFO)
	fmt.Println("  step 4 — function body done, deferred calls start")
}

// -----------------------------------------------------------------------
// SECTION 2: defer in a Loop
// -----------------------------------------------------------------------
// Deferred calls stack up in a loop — all run when the FUNCTION returns,
// not when the loop iteration ends. This is a common source of bugs.

func deferLoop() {
	fmt.Println("\ndefer in a loop (all fire at function end):")

	for i := 1; i <= 3; i++ {
		// Each defer captures the CURRENT value of i
		// (because i is passed as an argument — evaluated immediately)
		defer fmt.Printf("  deferred from loop i=%d\n", i)
	}
	fmt.Println("  loop done")
}

// -----------------------------------------------------------------------
// SECTION 3: defer Argument Evaluation
// -----------------------------------------------------------------------
// Arguments are evaluated when defer is called, NOT when the deferred
// function actually runs. This is a subtle but important distinction.

func deferArgEvaluation() {
	fmt.Println("\ndefer argument evaluation:")

	x := 1
	defer fmt.Printf("  deferred x = %d (captured at defer time = 1)\n", x)
	x = 100
	fmt.Printf("  x is now = %d\n", x)
	// deferred call runs with x=1, not x=100
}

// -----------------------------------------------------------------------
// SECTION 4: defer with Closures — Capturing by Reference
// -----------------------------------------------------------------------
// If you use a closure (anonymous function with no arguments), it
// captures variables by REFERENCE, reading them at execution time.

func deferClosure() {
	fmt.Println("\ndefer closure vs argument:")

	x := 1
	// Closure — captures x by reference, reads at execution time
	defer func() {
		fmt.Printf("  closure sees x = %d (at execution time = 100)\n", x)
	}()
	x = 100
	fmt.Println("  x is now 100")
}

// -----------------------------------------------------------------------
// SECTION 5: defer Modifying Named Return Values
// -----------------------------------------------------------------------
// A deferred function can read and modify the caller's named return values.
// This is the ONLY way a deferred function can change what the function returns.

func withCleanup() (result string, err error) {
	defer func() {
		// Wrap any error that occurred
		if err != nil {
			err = fmt.Errorf("withCleanup failed: %w", err)
			result = "error"
		}
	}()

	// Simulate some work
	result = "success"
	return result, nil
}

func deferNamedReturns() {
	fmt.Println("\ndefer with named returns:")
	r, err := withCleanup()
	fmt.Printf("  result=%q err=%v\n", r, err)
}

// -----------------------------------------------------------------------
// SECTION 6: Panic
// -----------------------------------------------------------------------
// panic() immediately stops the current function, unwinds the call stack
// running deferred functions, and crashes the program (unless recovered).
//
// Use panic ONLY for truly unrecoverable situations:
//   - Programmer errors (nil pointer, index out of bounds)
//   - Violated invariants that should never happen
//
// Do NOT use panic for normal error flow — use error returns instead.

func mustParsePositive(n int) int {
	if n <= 0 {
		// panic with a descriptive message
		panic(fmt.Sprintf("mustParsePositive: n must be > 0, got %d", n))
	}
	return n
}

func panicDemo() {
	fmt.Println("\npanic (deferred functions still run):")
	defer fmt.Println("  defer ran even during panic")

	// This panic would crash the program — we'll show recover next
	// mustParsePositive(-1)
	fmt.Println("  (panic example commented out to not crash)")
}

// -----------------------------------------------------------------------
// SECTION 7: recover — Catching Panics
// -----------------------------------------------------------------------
// recover() stops the panic and returns the value passed to panic().
// It ONLY works inside a deferred function.
// If no panic is happening, recover() returns nil.

// safeDiv divides a/b and recovers from a divide-by-zero panic
func safeDiv(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			// r is whatever was passed to panic()
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()
	return a / b, nil // integer division — panics if b==0
}

// safeRun runs any function and converts panics to errors
func safeRun(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	fn()
	return nil
}

func recoverDemo() {
	fmt.Println("\nrecover from panic:")

	result, err := safeDiv(10, 2)
	fmt.Printf("  10/2 = %d, err=%v\n", result, err)

	result, err = safeDiv(10, 0)
	fmt.Printf("  10/0 = %d, err=%v\n", result, err)

	// Recover from a panic caused by calling our function with bad input
	err = safeRun(func() {
		mustParsePositive(-5) // will panic
	})
	fmt.Printf("  safeRun panicking fn: err=%v\n", err)

	err = safeRun(func() {
		_ = mustParsePositive(5) // no panic
	})
	fmt.Printf("  safeRun normal fn:    err=%v\n", err)
}

// -----------------------------------------------------------------------
// SECTION 8: Real-World Pattern — Resource Cleanup
// -----------------------------------------------------------------------
// The most common use of defer: ensure resources are released.
// This pattern guarantees cleanup even if the function returns early or panics.

type Resource struct{ name string }

func (r *Resource) Open() error {
	fmt.Printf("    opened %s\n", r.name)
	return nil
}
func (r *Resource) Close() {
	fmt.Printf("    closed %s\n", r.name)
}
func (r *Resource) Process() error {
	fmt.Printf("    processing %s\n", r.name)
	return nil
}

func processWithCleanup(name string) error {
	r := &Resource{name: name}

	if err := r.Open(); err != nil {
		return err
	}
	defer r.Close() // guaranteed to run when processWithCleanup returns

	// Even if Process() errors or panics, r.Close() will run
	if err := r.Process(); err != nil {
		return err
	}

	return nil
}

func realWorldDefer() {
	fmt.Println("\nreal-world defer (resource cleanup):")
	processWithCleanup("database connection")
	processWithCleanup("file handle")
}

func main() {
	deferBasics()
	deferLoop()
	deferArgEvaluation()
	deferClosure()
	deferNamedReturns()
	panicDemo()
	recoverDemo()
	realWorldDefer()
}
