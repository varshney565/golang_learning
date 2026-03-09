// Package main demonstrates profiling with pprof in Go.
// Topics: CPU profile, memory profile, goroutine profile, flame graphs, HTTP pprof.
//
// pprof is the #1 performance tool in Go — asked in senior interviews.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: What is pprof?
// -----------------------------------------------------------------------
// pprof is Go's built-in profiler. It samples your program to show:
//
//   CPU profile     → which functions are consuming CPU time
//   Memory profile  → which code is allocating the most memory
//   Goroutine profile → which goroutines exist and their stack traces
//   Block profile   → where goroutines block on channel/sync ops
//   Mutex profile   → where mutex contention is happening
//   Trace          → detailed execution trace (go tool trace)
//
// Two ways to profile:
//   1. File-based: write profile to a file, analyze with: go tool pprof cpu.prof
//   2. HTTP-based: expose /debug/pprof/ endpoints (for live servers)

// -----------------------------------------------------------------------
// SECTION 2: CPU Profiling — File Based
// -----------------------------------------------------------------------

// cpuIntensiveWork simulates CPU-heavy computation
func cpuIntensiveWork(n int) float64 {
	result := 0.0
	for i := 0; i < n; i++ {
		// Expensive: lots of string operations
		s := strings.Repeat("a", 100)
		s = strings.ToUpper(s)
		result += float64(len(s))
	}
	return result
}

// anotherHotFunc simulates another hot path
func anotherHotFunc(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n * n
	}
	return total
}

func cpuProfilingDemo() {
	fmt.Println("CPU Profiling:")

	// Create profile output file
	f, err := os.CreateTemp("", "cpu_profile_*.prof")
	if err != nil {
		fmt.Printf("  could not create profile file: %v\n", err)
		return
	}
	defer os.Remove(f.Name()) // cleanup demo file

	fmt.Printf("  writing CPU profile to: %s\n", f.Name())

	// START profiling
	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Printf("  could not start CPU profile: %v\n", err)
		return
	}

	// ── Code to profile ──
	_ = cpuIntensiveWork(5000)
	nums := make([]int, 10000)
	for i := range nums {
		nums[i] = rand.Intn(100)
	}
	_ = anotherHotFunc(nums)
	// ── End of code to profile ──

	// STOP profiling — must call this before analyzing
	pprof.StopCPUProfile()
	f.Close()

	fmt.Println("  profile written. Analyze with:")
	fmt.Printf("    go tool pprof %s\n", f.Name())
	fmt.Println("  In the pprof shell:")
	fmt.Println("    top10          → show top 10 functions by CPU")
	fmt.Println("    list funcName  → annotated source of a function")
	fmt.Println("    web            → open flame graph in browser")
	fmt.Println("    svg            → generate SVG graph")
}

// -----------------------------------------------------------------------
// SECTION 3: Memory Profiling
// -----------------------------------------------------------------------

// memoryHog allocates a lot of memory
func memoryHog() [][]byte {
	result := make([][]byte, 0, 1000)
	for i := 0; i < 1000; i++ {
		// Allocate 1KB per item
		buf := make([]byte, 1024)
		rand.Read(buf)
		result = append(result, buf)
	}
	return result
}

func memoryProfilingDemo() {
	fmt.Println("\nMemory Profiling:")

	// Do some allocations
	data := memoryHog()
	_ = data

	// Write heap profile
	f, err := os.CreateTemp("", "mem_profile_*.prof")
	if err != nil {
		fmt.Printf("  could not create file: %v\n", err)
		return
	}
	defer os.Remove(f.Name())

	// Force GC to get accurate live heap data
	runtime.GC()

	// Write heap profile — captures current heap allocations
	if err := pprof.WriteHeapProfile(f); err != nil {
		fmt.Printf("  could not write heap profile: %v\n", err)
		return
	}
	f.Close()

	fmt.Printf("  heap profile written to: %s\n", f.Name())
	fmt.Println("  Analyze with:")
	fmt.Printf("    go tool pprof -alloc_space %s\n", f.Name())
	fmt.Println("  Flags:")
	fmt.Println("    -alloc_space  → total bytes allocated (ever)")
	fmt.Println("    -alloc_objects → total objects allocated (ever)")
	fmt.Println("    -inuse_space  → currently live bytes")
	fmt.Println("    -inuse_objects → currently live objects")
}

// -----------------------------------------------------------------------
// SECTION 4: Goroutine Profile
// -----------------------------------------------------------------------

func goroutineProfilingDemo() {
	fmt.Println("\nGoroutine Profile:")

	// Launch some goroutines
	for i := 0; i < 5; i++ {
		go func() {
			time.Sleep(100 * time.Millisecond) // simulate waiting goroutines
		}()
	}
	time.Sleep(10 * time.Millisecond) // let them start

	// Write goroutine profile to stdout
	fmt.Printf("  current goroutines: %d\n", runtime.NumGoroutine())
	fmt.Println("  goroutine profile (top stacks):")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1) // debug=1 shows counts
}

// -----------------------------------------------------------------------
// SECTION 5: HTTP pprof — Live Server Profiling
// -----------------------------------------------------------------------
// The most common way to profile production services.
// Just import the package — it registers /debug/pprof/ handlers automatically.

// ┌─────────────────────────────────────────────────────────────────────┐
// │ In your server code (usually main.go):                              │
// │                                                                     │
// │   import _ "net/http/pprof"  // registers /debug/pprof handlers    │
// │                                                                     │
// │   // Start a separate profiling server (never expose to public!)    │
// │   go http.ListenAndServe("localhost:6060", nil)                     │
// │                                                                     │
// │ Then from your terminal:                                            │
// │   # 30-second CPU profile                                           │
// │   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30│
// │                                                                     │
// │   # Heap snapshot                                                   │
// │   go tool pprof http://localhost:6060/debug/pprof/heap             │
// │                                                                     │
// │   # Goroutine dump                                                  │
// │   curl localhost:6060/debug/pprof/goroutine?debug=2                │
// │                                                                     │
// │ Interactive flame graph (requires graphviz):                        │
// │   go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap │
// └─────────────────────────────────────────────────────────────────────┘

func httpPprofExplained() {
	fmt.Println("\nHTTP pprof endpoints (after importing _ \"net/http/pprof\"):")
	endpoints := []struct{ path, desc string }{
		{"/debug/pprof/", "index page listing all profiles"},
		{"/debug/pprof/cmdline", "command line of the running program"},
		{"/debug/pprof/profile?seconds=30", "30-second CPU profile"},
		{"/debug/pprof/heap", "heap memory allocations"},
		{"/debug/pprof/goroutine", "all goroutines and their stacks"},
		{"/debug/pprof/block", "goroutine blocking on sync/channel"},
		{"/debug/pprof/mutex", "mutex contention"},
		{"/debug/pprof/trace?seconds=5", "execution trace for go tool trace"},
	}
	for _, e := range endpoints {
		fmt.Printf("  %-50s → %s\n", e.path, e.desc)
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Benchmark-Based Profiling
// -----------------------------------------------------------------------
// The most precise way: profile during a benchmark.
//
//   go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./...
//   go tool pprof cpu.prof
//
// Then in pprof:
//   top10
//   list FunctionName
//   web              (requires graphviz)
//
// Key metrics to look for:
//   flat%  → % of CPU this function directly consumed
//   cum%   → % of CPU this function + all callees consumed
//   alloc_space → total bytes allocated by this code path

func profilingWorkflow() {
	fmt.Println("\nProfiling workflow:")
	steps := []string{
		"1. Add profiling: import _ \"net/http/pprof\" or use file-based",
		"2. Reproduce the performance issue under profiling",
		"3. Analyze: go tool pprof -http=:8080 profile.prof",
		"4. Look at top functions in CPU profile (flat vs cum)",
		"5. Look at alloc_space in heap profile",
		"6. Focus on the HOTTEST path — don't micro-optimize cold code",
		"7. Make ONE change, benchmark before/after, repeat",
	}
	for _, s := range steps {
		fmt.Printf("  %s\n", s)
	}
}

// -----------------------------------------------------------------------
// SECTION 7: go test Benchmarks with Memory Stats
// -----------------------------------------------------------------------
// Run:  go test -bench=. -benchmem ./...
// Output shows:
//   BenchmarkFoo-8   1000000   1234 ns/op   128 B/op   2 allocs/op
//                              ^^^          ^^^         ^^^
//                         time per call   bytes     allocations

func main() {
	cpuProfilingDemo()
	memoryProfilingDemo()
	goroutineProfilingDemo()
	httpPprofExplained()
	profilingWorkflow()
}
