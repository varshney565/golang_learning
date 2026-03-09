// Package main demonstrates common concurrency patterns in Go.
// Topics: worker pool, fan-out/fan-in, pipeline, rate limiting.
package main

import (
	"fmt"
	"sync"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Worker Pool
// -----------------------------------------------------------------------
// Spawn N workers that pull jobs from a shared jobs channel.
// Limits concurrency to exactly N goroutines — prevents resource exhaustion.
//
//                 ┌──────────┐
//   jobs ──────>  │ worker 1 │ ──────> results
//                 │ worker 2 │ ──────> results
//                 │ worker 3 │ ──────> results
//                 └──────────┘

type Job struct {
	ID    int
	Input int
}

type Result struct {
	JobID  int
	Output int
}

func workerPool() {
	fmt.Println("Worker Pool:")

	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// Start N workers — each pulls from the jobs channel
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		w := w
		go func() {
			defer wg.Done()
			for job := range jobs { // blocks until job available or channel closed
				// Simulate work
				time.Sleep(time.Duration(job.Input) * time.Millisecond)
				results <- Result{JobID: job.ID, Output: job.Input * job.Input}
				fmt.Printf("  worker %d processed job %d\n", w, job.ID)
			}
		}()
	}

	// Send all jobs
	for i := 1; i <= numJobs; i++ {
		jobs <- Job{ID: i, Input: i}
	}
	close(jobs) // signal workers: no more jobs coming

	// Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var total int
	for r := range results {
		total += r.Output
	}
	fmt.Printf("  sum of squares 1..10 = %d\n", total)
}

// -----------------------------------------------------------------------
// SECTION 2: Fan-Out / Fan-In
// -----------------------------------------------------------------------
// Fan-out: distribute work from one channel to multiple workers.
// Fan-in: collect results from multiple workers into one channel.
//
//              ┌──> worker A ──┐
//   input ──>  ├──> worker B ──┼──> merged output
//              └──> worker C ──┘

func fanOutFanIn() {
	fmt.Println("\nFan-Out / Fan-In:")

	// Source of numbers to process
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	in := make(chan int, len(nums))
	for _, n := range nums {
		in <- n
	}
	close(in)

	// Fan-out: launch N workers, each reading from the same input channel
	const numWorkers = 3
	workerChans := make([]<-chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		out := make(chan int)
		workerChans[i] = out
		go func(out chan<- int) {
			for n := range in {
				out <- n * n // square each number
			}
			close(out)
		}(out)
	}

	// Fan-in: merge all worker outputs into one channel
	merged := merge(workerChans...)

	// Collect all results
	var results []int
	for v := range merged {
		results = append(results, v)
	}
	fmt.Printf("  results: %v\n", results)
}

// merge takes any number of channels and merges them into one
func merge(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	// Forward from each input channel to output
	forward := func(ch <-chan int) {
		defer wg.Done()
		for v := range ch {
			out <- v
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go forward(ch)
	}

	// Close output when all inputs are exhausted
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// -----------------------------------------------------------------------
// SECTION 3: Pipeline with Cancellation
// -----------------------------------------------------------------------
// A real-world pipeline with done-channel cancellation.
// Any stage can stop the whole pipeline by closing done.

func pipelineWithCancellation() {
	fmt.Println("\nPipeline with cancellation:")

	done := make(chan struct{})
	defer close(done) // ensures all goroutines are cleaned up when main returns

	// Stage 1: generate numbers
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for _, n := range nums {
				select {
				case out <- n:
				case <-done: // stop if cancelled
					return
				}
			}
		}()
		return out
	}

	// Stage 2: square
	sq := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				select {
				case out <- n * n:
				case <-done:
					return
				}
			}
		}()
		return out
	}

	// Stage 3: filter (keep > 10)
	filter := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				if n > 10 {
					select {
					case out <- n:
					case <-done:
						return
					}
				}
			}
		}()
		return out
	}

	// Wire up: gen → sq → filter → consume
	nums := gen(1, 2, 3, 4, 5)
	squares := sq(nums)
	filtered := filter(squares)

	for v := range filtered {
		fmt.Printf("  %d\n", v) // only squares > 10: 16, 25
	}
}

// -----------------------------------------------------------------------
// SECTION 4: Rate Limiter
// -----------------------------------------------------------------------
// Use time.Tick (or time.NewTicker) to limit the rate of operations.

func rateLimiter() {
	fmt.Println("\nRate limiter (3 requests/sec):")

	requests := []int{1, 2, 3, 4, 5}
	// Allow 1 request per 100ms (= ~10/sec), limit to 3 with burst
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	for _, req := range requests {
		<-limiter.C // wait for tick
		fmt.Printf("  request %d at %v\n", req, time.Now().Format("15:04:05.000"))
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Errgroup Pattern
// -----------------------------------------------------------------------
// Similar to WaitGroup but collects errors from goroutines.
// Uses golang.org/x/sync/errgroup in real projects; shown here manually.

type errGroup struct {
	wg   sync.WaitGroup
	once sync.Once
	err  error
}

func (eg *errGroup) Go(fn func() error) {
	eg.wg.Add(1)
	go func() {
		defer eg.wg.Done()
		if err := fn(); err != nil {
			// Only store the FIRST error (like errgroup does)
			eg.once.Do(func() { eg.err = err })
		}
	}()
}

func (eg *errGroup) Wait() error {
	eg.wg.Wait()
	return eg.err
}

func errGroupDemo() {
	fmt.Println("\nErrGroup pattern:")

	var eg errGroup

	eg.Go(func() error {
		return nil // success
	})
	eg.Go(func() error {
		return fmt.Errorf("task 2 failed")
	})
	eg.Go(func() error {
		return fmt.Errorf("task 3 also failed")
	})

	if err := eg.Wait(); err != nil {
		fmt.Printf("  first error: %v\n", err)
	}
}

func main() {
	workerPool()
	fanOutFanIn()
	pipelineWithCancellation()
	rateLimiter()
	errGroupDemo()
}
