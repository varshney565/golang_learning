// Package main demonstrates the sync package in Go.
// Topics: Mutex, RWMutex, WaitGroup, Once, atomic operations, Map.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// -----------------------------------------------------------------------
// SECTION 1: sync.Mutex — Mutual Exclusion
// -----------------------------------------------------------------------
// A Mutex ensures only ONE goroutine accesses a critical section at a time.
//
//   mu.Lock()    → acquire the lock (blocks if already held)
//   mu.Unlock()  → release the lock
//
// Always unlock in a defer to ensure release even on panic.

type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock() // always use defer for Unlock
	c.count++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func mutexDemo() {
	fmt.Println("Mutex:")

	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()
	fmt.Printf("  counter = %d (expected 1000)\n", counter.Value())
}

// -----------------------------------------------------------------------
// SECTION 2: sync.RWMutex — Multiple Readers, One Writer
// -----------------------------------------------------------------------
// RWMutex is more efficient when reads are far more common than writes.
//
//   mu.RLock() / mu.RUnlock()  → multiple goroutines can hold read lock
//   mu.Lock()  / mu.Unlock()   → exclusive write lock (no readers or writers)
//
// Use RWMutex for read-heavy caches, config, registries.

type SafeCache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewSafeCache() *SafeCache {
	return &SafeCache{data: make(map[string]string)}
}

func (c *SafeCache) Set(key, value string) {
	c.mu.Lock() // exclusive write lock
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *SafeCache) Get(key string) (string, bool) {
	c.mu.RLock() // shared read lock — allows concurrent reads
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func rwMutexDemo() {
	fmt.Println("\nRWMutex (cache):")

	cache := NewSafeCache()
	var wg sync.WaitGroup

	// 1 writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set("host", "localhost")
		cache.Set("port", "8080")
	}()

	wg.Wait() // ensure writes are done before reads

	// 5 concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if host, ok := cache.Get("host"); ok {
				_ = host // use value
			}
		}()
	}
	wg.Wait()

	host, _ := cache.Get("host")
	fmt.Printf("  cache[host] = %q\n", host)
}

// -----------------------------------------------------------------------
// SECTION 3: sync.WaitGroup — Goroutine Synchronization
// -----------------------------------------------------------------------
// Already covered in goroutines, but here's a more detailed look.

func waitGroupDemo() {
	fmt.Println("\nWaitGroup:")

	var wg sync.WaitGroup
	results := make([]int, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			results[i] = i * i // safe: each goroutine writes its own index
		}()
	}

	wg.Wait()
	fmt.Printf("  results: %v\n", results)
}

// -----------------------------------------------------------------------
// SECTION 4: sync.Once — Run Exactly Once
// -----------------------------------------------------------------------
// Once.Do(f) ensures f is called exactly once, even across goroutines.
// Used for lazy initialization, singleton pattern.

type Database struct {
	connection string
}

var (
	dbInstance *Database
	dbOnce     sync.Once
)

func getDatabase() *Database {
	dbOnce.Do(func() {
		// This runs only once, no matter how many goroutines call getDatabase()
		fmt.Println("  initializing database (runs once)")
		dbInstance = &Database{connection: "db://localhost:5432"}
	})
	return dbInstance
}

func onceDemo() {
	fmt.Println("\nsync.Once (singleton):")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db := getDatabase()
			_ = db
		}()
	}
	wg.Wait()
	fmt.Printf("  db connection: %s\n", getDatabase().connection)
}

// -----------------------------------------------------------------------
// SECTION 5: sync/atomic — Lock-Free Primitives
// -----------------------------------------------------------------------
// atomic operations are guaranteed to be indivisible (atomic).
// Much faster than mutexes for simple counters, flags, and pointers.
//
// Key functions:
//   atomic.AddInt64(&v, delta)   → add and return new value
//   atomic.LoadInt64(&v)         → read atomically
//   atomic.StoreInt64(&v, n)     → write atomically
//   atomic.CompareAndSwapInt64   → CAS: write only if current value matches

func atomicDemo() {
	fmt.Println("\natomic operations:")

	var counter int64 // must use int32 or int64 with atomic
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1) // atomic increment — no mutex needed
		}()
	}
	wg.Wait()
	fmt.Printf("  counter = %d (expected 1000)\n", atomic.LoadInt64(&counter))

	// Compare-and-Swap (CAS) — update only if current value matches expected
	// Useful for lock-free algorithms
	var flag int32 = 0
	swapped := atomic.CompareAndSwapInt32(&flag, 0, 1) // set to 1 if currently 0
	fmt.Printf("  CAS (0→1): swapped=%v, flag=%d\n", swapped, flag)

	swapped = atomic.CompareAndSwapInt32(&flag, 0, 1) // fails — flag is now 1, not 0
	fmt.Printf("  CAS (0→1): swapped=%v, flag=%d (already 1)\n", swapped, flag)
}

// -----------------------------------------------------------------------
// SECTION 6: sync.Map — Concurrent-Safe Map
// -----------------------------------------------------------------------
// sync.Map is a built-in concurrent-safe map.
// Use it when:
//   - The key set is mostly stable (written once, read many times)
//   - Multiple goroutines read/write different keys (no contention)
//
// For general use, a regular map + RWMutex is often clearer and equally fast.

func syncMapDemo() {
	fmt.Println("\nsync.Map:")

	var m sync.Map
	var wg sync.WaitGroup

	// Concurrent writes (different keys — no contention)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			m.Store(fmt.Sprintf("key%d", i), i*10)
		}()
	}
	wg.Wait()

	// Load values
	for i := 0; i < 5; i++ {
		if v, ok := m.Load(fmt.Sprintf("key%d", i)); ok {
			fmt.Printf("  key%d = %v\n", i, v)
		}
	}

	// Range over all entries
	fmt.Println("  all entries:")
	m.Range(func(k, v any) bool {
		fmt.Printf("    %v → %v\n", k, v)
		return true // return false to stop iteration
	})
}

// -----------------------------------------------------------------------
// SECTION 7: sync.Pool — Object Reuse Pool
// -----------------------------------------------------------------------
// sync.Pool reduces GC pressure by reusing objects.
// Objects in the pool may be GC'd at any time — don't store state in them.
// Common use: reusing byte buffers, encoders, large temporary objects.

func poolDemo() {
	fmt.Println("\nsync.Pool:")

	pool := &sync.Pool{
		New: func() any {
			fmt.Println("    allocating new buffer")
			return make([]byte, 1024)
		},
	}

	// Get from pool (or allocate via New)
	buf := pool.Get().([]byte)
	fmt.Printf("  got buffer of len %d\n", len(buf))

	// Use the buffer...
	copy(buf, "hello world")

	// Return to pool for reuse (reset state first!)
	buf = buf[:0]        // reset length but keep backing array
	pool.Put(buf[:1024]) // put back full capacity

	// Next Get() may return the pooled buffer or allocate a new one
	buf2 := pool.Get().([]byte)
	fmt.Printf("  got buffer again, len %d\n", len(buf2))
}

func main() {
	mutexDemo()
	rwMutexDemo()
	waitGroupDemo()
	onceDemo()
	atomicDemo()
	syncMapDemo()
	poolDemo()
}
