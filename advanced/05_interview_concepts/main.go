// Package main covers concepts that come up in Go interviews.
// Topics: goroutine scheduler, channel internals, string internals,
//         interface internals, init order, common gotchas.
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// -----------------------------------------------------------------------
// SECTION 1: Go Scheduler (GMP Model) — Conceptual
// -----------------------------------------------------------------------
// Go uses the GMP scheduler (Goroutine, Machine, Processor):
//
//   G (Goroutine) — lightweight thread of execution (~2KB stack, grows as needed)
//   M (Machine)   — OS thread
//   P (Processor) — logical CPU, holds a queue of goroutines to run
//
// Default: GOMAXPROCS P's, one per CPU core
//
// When a goroutine blocks on I/O:
//   The OS thread (M) is detached from the processor (P).
//   P picks up another goroutine to keep running.
//   When I/O completes, goroutine goes back into a run queue.
//
// This is why Go can have millions of goroutines — blocked ones don't
// hold OS threads.
//
// Work stealing: idle P's steal goroutines from busy P's run queues.

func schedulerConcept() {
	fmt.Println("GMP Scheduler (conceptual):")
	fmt.Println("  G = Goroutine (lightweight, ~2KB stack)")
	fmt.Println("  M = Machine (OS thread)")
	fmt.Println("  P = Processor (logical CPU, runs goroutines)")
	fmt.Println("  Goroutines blocking on I/O don't hold OS threads")
	fmt.Println("  Work stealing keeps all P's busy")
}

// -----------------------------------------------------------------------
// SECTION 2: String Internals
// -----------------------------------------------------------------------
// A string in Go is an IMMUTABLE sequence of bytes (not characters).
// It's a struct: { pointer to byte array, length }
//
// Strings can contain any bytes — they're UTF-8 by CONVENTION.
// A "rune" (int32) represents a Unicode code point.

func stringInternals() {
	fmt.Println("\nString internals:")

	s := "Hello, 世界" // "世界" is Chinese — each char is 3 bytes in UTF-8

	fmt.Printf("  s = %q\n", s)
	fmt.Printf("  len(s) = %d (byte count, NOT char count)\n", len(s))
	fmt.Printf("  utf8.RuneCountInString = %d (actual character count)\n", utf8.RuneCountInString(s))

	// Byte indexing — can split multi-byte runes if not careful
	fmt.Printf("  s[0] = %d (byte value of 'H')\n", s[0])

	// Ranging over a string iterates RUNES (Unicode code points)
	fmt.Println("  range iterates runes:")
	for i, r := range s {
		if i < 10 {
			fmt.Printf("    index=%d rune=%c (%d)\n", i, r, r)
		}
	}

	// Convert to []rune for character-level operations
	runes := []rune(s)
	fmt.Printf("  runes[7] = %c (correct char access)\n", runes[7])

	// String immutability — you can't modify s[0] = 'h'
	// Must convert to []byte first
	b := []byte(s)
	b[0] = 'h'
	fmt.Printf("  modified: %q\n", string(b))

	// strings.Builder for efficient concatenation
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		sb.WriteString(fmt.Sprintf("item%d ", i))
	}
	fmt.Printf("  builder: %q\n", sb.String())
}

// -----------------------------------------------------------------------
// SECTION 3: Interface Internals
// -----------------------------------------------------------------------
// An interface value has two components:
//   (type, value) — a type pointer and a data pointer
//
// nil interface: both type AND value are nil
// Non-nil interface with nil value: type is non-nil, value is nil
//    This is the famous "nil interface gotcha" (covered in phase2/02)

type Animal interface{ Sound() string }
type Dog struct{ Name string }
func (d *Dog) Sound() string { return "woof" }

func interfaceInternals() {
	fmt.Println("\nInterface internals:")

	var a Animal    // nil interface — both type and value are nil
	fmt.Printf("  nil interface: %v, isNil=%v\n", a, a == nil)

	var d *Dog      // nil pointer
	a = d           // interface now holds (type=*Dog, value=nil)
	// a != nil because the TYPE part is non-nil!
	fmt.Printf("  (*Dog)(nil): %v, isNil=%v ← GOTCHA\n", a, a == nil)

	// Non-nil value
	a = &Dog{Name: "Rex"}
	fmt.Printf("  &Dog{}: %v, sound=%s\n", a, a.Sound())

	// Interface is fat pointer: {type_ptr, data_ptr}
	// Small values (≤ pointer size) are stored inline
	// Large values are heap-allocated, interface holds a pointer
}

// -----------------------------------------------------------------------
// SECTION 4: Common Interview Gotchas
// -----------------------------------------------------------------------

// GOTCHA 1: Range loop variable reuse (Go < 1.22)
func rangeGotcha() {
	fmt.Println("\nRange gotcha (Go < 1.22):")

	// BUG: all goroutines capture same `v` variable (its final value)
	nums := []int{1, 2, 3}
	funcs := make([]func(), 3)
	for i, v := range nums {
		i, v := i, v // FIX: shadow with new variable (not needed in Go 1.22+)
		funcs[i] = func() { fmt.Printf("  %d", v) }
	}
	for _, f := range funcs {
		f()
	}
	fmt.Println()
}

// GOTCHA 2: Slice capacity sharing
func sliceCapGotcha() {
	fmt.Println("\nSlice cap gotcha:")

	a := []int{1, 2, 3, 4, 5}
	b := a[:3] // shares backing array, cap=5

	// append to b — fits within cap, modifies a's elements!
	b = append(b, 99)
	fmt.Printf("  a after append to b: %v ← a[3] changed!\n", a)

	// Fix: use full slice expression to set cap, forcing new allocation
	c := a[:3:3] // cap explicitly = 3
	c = append(c, 99)
	fmt.Printf("  a after append to c: %v ← a unchanged\n", a)
}

// GOTCHA 3: Map is not safe for concurrent access
func mapConcurrencyGotcha() {
	fmt.Println("\nMap concurrency gotcha:")
	fmt.Println("  maps are NOT safe for concurrent read+write")
	fmt.Println("  Use sync.RWMutex or sync.Map")
	fmt.Println("  Go runtime will panic/throw on concurrent map writes")
}

// GOTCHA 4: Defer in loop
func deferInLoopGotcha() {
	fmt.Println("\nDefer in loop gotcha:")
	fmt.Println("  Defers inside a loop pile up — run at FUNCTION end, not loop end")
	fmt.Println("  Fix: wrap loop body in an anonymous function")

	// BAD: all defers pile up, files would be left open until function returns
	// for _, f := range files {
	//     fh, _ := os.Open(f)
	//     defer fh.Close()  ← piles up
	// }

	// GOOD: scope defer to the inner function
	process := func(name string) {
		fmt.Printf("  process(%s): open and close within this func\n", name)
		// fh, _ := os.Open(name)
		// defer fh.Close()  ← runs when THIS func returns
	}
	for _, name := range []string{"a", "b", "c"} {
		process(name)
	}
}

// GOTCHA 5: Named return gotcha with defer
func namedReturnGotcha() (result int) {
	defer func() {
		result++ // modifies the named return value!
	}()
	return 0 // sets result=0, then defer increments it to 1
}

// GOTCHA 6: goroutine in test
// When a test function returns, all goroutines it launched are killed.
// Use WaitGroup or channels to wait for goroutines in tests.

// -----------------------------------------------------------------------
// SECTION 5: Make vs New
// -----------------------------------------------------------------------
// new(T)        — allocates zero-value T, returns *T
//                 Works for any type
//
// make(T, args) — allocates and INITIALIZES the type
//                 Only works for: slice, map, chan
//                 Returns T (not *T)

func makeVsNew() {
	fmt.Println("\nmake vs new:")

	// new — returns pointer to zero value
	pi := new(int)
	fmt.Printf("  new(int):   *pi = %d\n", *pi)

	ps := new([]int) // *[]int — pointer to nil slice
	fmt.Printf("  new([]int): *ps = %v, nil=%v\n", *ps, *ps == nil)

	// make — returns initialized value (not pointer)
	s := make([]int, 3, 5)   // initialized slice
	m := make(map[string]int) // initialized map
	c := make(chan int, 1)    // initialized channel

	fmt.Printf("  make([]int,3,5):      %v len=%d cap=%d\n", s, len(s), cap(s))
	fmt.Printf("  make(map[string]int): %v nil=%v\n", m, m == nil)
	fmt.Printf("  make(chan int, 1):    %v\n", c)
}

// -----------------------------------------------------------------------
// SECTION 6: When to Use Each Concurrency Primitive
// -----------------------------------------------------------------------

func concurrencyChoice() {
	fmt.Println("\nWhich concurrency primitive to use:")
	choices := []struct{ condition, answer string }{
		{"Transfer data between goroutines", "channel"},
		{"Signal completion (1 goroutine)", "close a channel or channel<-struct{}{}"},
		{"Broadcast stop to many goroutines", "close a done channel"},
		{"Protect shared state", "sync.Mutex (or sync.RWMutex for read-heavy)"},
		{"Wait for N goroutines to finish", "sync.WaitGroup"},
		{"Run something exactly once", "sync.Once"},
		{"Simple counter/flag", "sync/atomic"},
		{"Context propagation + cancellation", "context.Context"},
		{"Limit concurrency to N", "buffered channel as semaphore"},
	}
	for _, c := range choices {
		fmt.Printf("  %-50s → %s\n", c.condition, c.answer)
	}
}

// -----------------------------------------------------------------------
// SECTION 7: Key Interview Questions
// -----------------------------------------------------------------------

func interviewQuestions() {
	fmt.Println("\nKey interview topics (know these deeply):")
	topics := []string{
		"What is a goroutine vs OS thread?",
		"Explain the GMP scheduler model",
		"What is a data race? How to detect and fix?",
		"Buffered vs unbuffered channels — when to use each?",
		"Explain interface nil gotcha",
		"How does defer work with named returns?",
		"Explain escape analysis (stack vs heap)",
		"How does Go's GC work? How to tune it?",
		"What is the difference between make and new?",
		"How do you implement a worker pool?",
		"Explain errors.Is vs errors.As",
		"What is a closure? What are common closure bugs?",
		"Explain embedding vs inheritance",
		"When would you use reflect?",
		"How do generics work in Go?",
		"What is context.Context and why is it important?",
		"How do you prevent goroutine leaks?",
	}
	for i, q := range topics {
		fmt.Printf("  %2d. %s\n", i+1, q)
	}
}

func main() {
	schedulerConcept()
	stringInternals()
	interfaceInternals()
	rangeGotcha()
	sliceCapGotcha()
	mapConcurrencyGotcha()
	deferInLoopGotcha()

	result := namedReturnGotcha()
	fmt.Printf("\nNamed return with defer: expected 1, got %d\n", result)

	makeVsNew()
	concurrencyChoice()
	interviewQuestions()
}
