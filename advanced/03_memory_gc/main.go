// Package main explains Go's memory model, GC, escape analysis, and optimization.
// This is a senior-level interview topic.
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"unsafe"
)

// -----------------------------------------------------------------------
// SECTION 1: Stack vs Heap
// -----------------------------------------------------------------------
// Go automatically decides where to allocate:
//
//   STACK: Fast allocation/deallocation. Limited size (~1GB by default).
//          Local variables that don't escape the function.
//
//   HEAP:  Slower (GC managed). Grows dynamically.
//          Variables that ESCAPE to heap (survive function return, etc.).
//
// The compiler's ESCAPE ANALYSIS determines which allocation to use.
// You can see it with: go build -gcflags="-m" ./...

// This value stays on the STACK — it doesn't escape
func stackAlloc() int {
	x := 42    // on stack
	return x   // value is copied to caller — x itself stays local
}

// This value ESCAPES to HEAP — pointer returned means it must outlive the func
func heapAlloc() *int {
	x := 42    // x escapes to heap (compiler detects this)
	return &x  // we return the address, so x must live beyond this function
}

// Slices larger than a threshold escape to heap
func largeSlice() []int {
	// This likely escapes to heap because it's passed out
	s := make([]int, 1000)
	return s
}

func stackVsHeap() {
	fmt.Println("Stack vs Heap:")
	sv := stackAlloc()
	hv := heapAlloc()
	fmt.Printf("  stack value: %d\n", sv)
	fmt.Printf("  heap value:  %d\n", *hv)
	fmt.Println("  (run: go build -gcflags=-m to see escape analysis)")
}

// -----------------------------------------------------------------------
// SECTION 2: Go's Garbage Collector (GC)
// -----------------------------------------------------------------------
// Go uses a CONCURRENT, TRI-COLOR MARK-AND-SWEEP GC.
//
// Key properties:
//   - Concurrent: GC runs concurrently with your program (low pause times)
//   - Tri-color mark-and-sweep:
//       White: not yet visited (candidates for collection)
//       Grey:  visited but children not yet scanned
//       Black: fully scanned (not garbage)
//   - Stop-The-World (STW) pauses are very short (< 1ms typically)
//
// GC runs when:
//   - Heap size doubles since last GC (default GOGC=100)
//   - Memory limit is approached (GOMEMLIMIT)
//   - runtime.GC() is called manually

func gcDemo() {
	fmt.Println("\nGarbage Collector:")

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("  HeapAlloc:   %d KB\n", stats.HeapAlloc/1024)
	fmt.Printf("  HeapSys:     %d KB\n", stats.HeapSys/1024)
	fmt.Printf("  NumGC:       %d\n", stats.NumGC)
	fmt.Printf("  GCCPUFrac:   %.4f (fraction of CPU used by GC)\n", stats.GCCPUFraction)

	// Allocate some objects
	data := make([][]byte, 1000)
	for i := range data {
		data[i] = make([]byte, 1024)
	}

	// Force GC (don't do this in production!)
	runtime.GC()
	data = nil // make data eligible for collection
	runtime.GC()

	runtime.ReadMemStats(&stats)
	fmt.Printf("  After GC — NumGC: %d\n", stats.NumGC)
}

// -----------------------------------------------------------------------
// SECTION 3: GOGC and Memory Tuning
// -----------------------------------------------------------------------
// GOGC env var controls GC trigger:
//   GOGC=100 (default): GC when heap doubles
//   GOGC=200: GC when heap triples (less frequent GC, more memory)
//   GOGC=50:  GC more frequently (less memory, more CPU)
//   GOGC=off: disable GC entirely (dangerous!)
//
// GOMEMLIMIT (Go 1.19+): hard memory limit
//   GOMEMLIMIT=500MiB: trigger GC before reaching 500MB
//   This is the recommended way to limit memory in containers

func gcTuning() {
	fmt.Println("\nGC Tuning:")

	// Get current GOGC setting
	prev := debug.SetGCPercent(-1) // -1 returns current without changing
	debug.SetGCPercent(prev)       // restore
	fmt.Printf("  GOGC setting (SetGCPercent): ~%d\n", prev)
	fmt.Println("  Set GOGC=off in production only for batch jobs")
	fmt.Println("  Use GOMEMLIMIT for container deployments")
}

// -----------------------------------------------------------------------
// SECTION 4: Memory Allocation Reduction Tips
// -----------------------------------------------------------------------

// TIP 1: Pre-allocate slices when you know the size
func preallocateSlice(n int) []int {
	result := make([]int, 0, n) // capacity=n avoids reallocations
	for i := 0; i < n; i++ {
		result = append(result, i)
	}
	return result
}

// TIP 2: Use value receivers for small structs (avoids heap allocation)
type SmallPoint struct{ X, Y float64 } // 16 bytes — pass by value
func (p SmallPoint) Magnitude() float64 {
	return p.X*p.X + p.Y*p.Y
}

// TIP 3: Reuse objects with sync.Pool (see phase3/04_sync)
// TIP 4: Avoid interface boxing of small values when in hot paths
// TIP 5: Use []byte instead of string for manipulation (strings are immutable = allocation)

func allocationTips() {
	fmt.Println("\nAllocation reduction tips:")

	// String concatenation in a loop — BAD (allocates each time)
	// result := ""
	// for i := 0; i < 1000; i++ { result += "x" }  // O(n^2) allocations

	// Use strings.Builder — GOOD
	// var sb strings.Builder
	// for i := 0; i < 1000; i++ { sb.WriteByte('x') }
	// result := sb.String()

	// Converting string to []byte once and back — better than repeated conversions
	s := "hello world"
	b := []byte(s)   // allocates — but only once
	b[0] = 'H'
	s2 := string(b) // allocates — but only once
	fmt.Printf("  string manipulation: %q\n", s2)

	_ = preallocateSlice(10)
	fmt.Println("  preallocate with make([]T, 0, n) to avoid reallocations")
}

// -----------------------------------------------------------------------
// SECTION 5: Memory Layout — Struct Field Ordering
// -----------------------------------------------------------------------
// Go does NOT automatically reorder struct fields.
// Misaligned fields cause padding, wasting memory.
// Order fields from largest to smallest to minimize padding.

type BadLayout struct {
	A bool    // 1 byte + 7 bytes padding
	B float64 // 8 bytes
	C bool    // 1 byte + 7 bytes padding
	// Total: 24 bytes
}

type GoodLayout struct {
	B float64 // 8 bytes
	A bool    // 1 byte
	C bool    // 1 byte + 6 bytes padding
	// Total: 16 bytes
}

func structLayout() {
	fmt.Println("\nStruct memory layout:")
	fmt.Printf("  BadLayout:  %d bytes\n", unsafe.Sizeof(BadLayout{}))
	fmt.Printf("  GoodLayout: %d bytes\n", unsafe.Sizeof(GoodLayout{}))
	fmt.Println("  Order fields large→small to minimize padding")
}

// -----------------------------------------------------------------------
// SECTION 6: False Sharing — CPU Cache Lines
// -----------------------------------------------------------------------
// Modern CPUs cache memory in 64-byte cache lines.
// If two goroutines write to different fields in the same cache line,
// they invalidate each other's cache — "false sharing".
// Solution: pad structs to align on cache line boundaries.

// CacheLinePad is 64 bytes (one CPU cache line)
type CacheLinePad [64]byte

// Without padding — goroutines writing A and B may cause false sharing
type CountersWithoutPad struct {
	A int64
	B int64 // shares cache line with A
}

// With padding — A and B are in separate cache lines
type CountersWithPad struct {
	A   int64
	_   CacheLinePad // force B to a different cache line
	B   int64
}

func cacheLinePadding() {
	fmt.Println("\nCache line padding:")
	fmt.Printf("  CountersWithoutPad: %d bytes\n", unsafe.Sizeof(CountersWithoutPad{}))
	fmt.Printf("  CountersWithPad:    %d bytes (padded for performance)\n", unsafe.Sizeof(CountersWithPad{}))
	fmt.Println("  Use padding in high-contention concurrent counters")
}

func main() {
	stackVsHeap()
	gcDemo()
	gcTuning()
	allocationTips()
	structLayout()
	cacheLinePadding()
}
