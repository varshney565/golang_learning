// Package main demonstrates the context package in Go.
// Topics: cancellation, deadlines, timeouts, values, propagation.
package main

import (
	"context"
	"fmt"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: What is context?
// -----------------------------------------------------------------------
// context.Context carries:
//   1. Cancellation signal  — stop working when told to
//   2. Deadline/Timeout     — stop working after a time
//   3. Values               — request-scoped data (trace IDs, auth tokens)
//
// Contexts form a TREE. Cancelling a parent also cancels all children.
// Pass context as the FIRST parameter in functions: func Foo(ctx context.Context, ...)
// This is a Go convention enforced by the community.

// -----------------------------------------------------------------------
// SECTION 2: Background and TODO
// -----------------------------------------------------------------------
// context.Background() — the root context, never cancelled, no deadline.
//   Use at the top of call chains (main, HTTP handler, test).
//
// context.TODO()       — placeholder when you're not sure which context to use.
//   Replace with proper context later.

func backgroundAndTODO() {
	bg := context.Background()
	todo := context.TODO()
	fmt.Printf("Background: %v\n", bg)
	fmt.Printf("TODO:       %v\n", todo)
}

// -----------------------------------------------------------------------
// SECTION 3: WithCancel — Manual Cancellation
// -----------------------------------------------------------------------
// WithCancel returns a copy of the context with a cancel function.
// Call cancel() to signal all goroutines using this context to stop.
// Always defer cancel() to prevent context leak.

func doWork(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			// ctx.Err() tells you WHY it was cancelled
			fmt.Printf("  %s: stopped (%v)\n", name, ctx.Err())
			return
		default:
			fmt.Printf("  %s: working...\n", name)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func withCancelDemo() {
	fmt.Println("\nWithCancel:")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // always defer — cleans up resources even on early return

	go doWork(ctx, "worker-A")
	go doWork(ctx, "worker-B")

	time.Sleep(120 * time.Millisecond) // let workers run a bit
	fmt.Println("  cancelling...")
	cancel() // signal both workers to stop

	time.Sleep(50 * time.Millisecond) // wait for goroutines to print stop message
}

// -----------------------------------------------------------------------
// SECTION 4: WithTimeout — Automatic Cancellation After Duration
// -----------------------------------------------------------------------
// WithTimeout(parent, d) cancels the context after duration d.
// Returns the context AND a cancel function (call it to release resources early).
// ctx.Err() == context.DeadlineExceeded when timeout fires.

func fetchData(ctx context.Context, url string) (string, error) {
	// Simulate a slow external call
	done := make(chan string, 1)
	go func() {
		time.Sleep(200 * time.Millisecond) // simulate latency
		done <- "data from " + url
	}()

	select {
	case data := <-done:
		return data, nil
	case <-ctx.Done():
		return "", fmt.Errorf("fetchData %s: %w", url, ctx.Err())
	}
}

func withTimeoutDemo() {
	fmt.Println("\nWithTimeout:")

	// Short timeout — will expire before fetchData finishes
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data, err := fetchData(ctx, "https://api.example.com")
	if err != nil {
		fmt.Printf("  error: %v\n", err)
	} else {
		fmt.Printf("  data: %s\n", data)
	}

	// Long timeout — fetchData finishes in time
	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()

	data, err = fetchData(ctx2, "https://fast.example.com")
	if err != nil {
		fmt.Printf("  error: %v\n", err)
	} else {
		fmt.Printf("  data: %s\n", data)
	}
}

// -----------------------------------------------------------------------
// SECTION 5: WithDeadline — Cancellation at a Specific Time
// -----------------------------------------------------------------------
// Like WithTimeout but takes an absolute time.Time instead of a duration.

func withDeadlineDemo() {
	fmt.Println("\nWithDeadline:")

	deadline := time.Now().Add(150 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	fmt.Printf("  deadline: %v\n", deadline.Format("15:04:05.000"))

	// Check how much time remains
	if d, ok := ctx.Deadline(); ok {
		fmt.Printf("  time until deadline: %v\n", time.Until(d).Round(time.Millisecond))
	}

	// Simulate work that exceeds deadline
	select {
	case <-time.After(300 * time.Millisecond):
		fmt.Println("  work done")
	case <-ctx.Done():
		fmt.Printf("  deadline exceeded: %v\n", ctx.Err())
	}
}

// -----------------------------------------------------------------------
// SECTION 6: WithValue — Passing Request-Scoped Values
// -----------------------------------------------------------------------
// WithValue stores a key-value pair in the context.
// Use for request-scoped data: trace IDs, auth tokens, user IDs.
//
// Rules:
//   1. Use unexported keys (not strings) to avoid collisions
//   2. Don't store optional parameters — use function args for those
//   3. Values should be small and immutable
//   4. Don't use for everything — only truly request-scoped data

// Use a custom unexported type as key to avoid collisions with other packages
type contextKey int

const (
	requestIDKey contextKey = iota
	userIDKey
)

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func getRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

func handler(ctx context.Context) {
	reqID, ok := getRequestID(ctx)
	if !ok {
		reqID = "unknown"
	}
	fmt.Printf("  handling request [%s]\n", reqID)
}

func withValueDemo() {
	fmt.Println("\nWithValue:")

	ctx := context.Background()
	ctx = withRequestID(ctx, "req-abc-123")
	ctx = context.WithValue(ctx, userIDKey, 42)

	handler(ctx)

	// Access values anywhere downstream
	if uid, ok := ctx.Value(userIDKey).(int); ok {
		fmt.Printf("  user ID: %d\n", uid)
	}

	// Missing key returns nil (not panics)
	missing := ctx.Value("nonexistent")
	fmt.Printf("  missing key: %v\n", missing)
}

// -----------------------------------------------------------------------
// SECTION 7: Context Propagation in a Call Chain
// -----------------------------------------------------------------------
// The real power: pass one context down through many layers.
// Cancellation propagates automatically.

func serviceA(ctx context.Context) error {
	fmt.Println("  serviceA: calling serviceB")
	return serviceB(ctx)
}

func serviceB(ctx context.Context) error {
	fmt.Println("  serviceB: calling serviceC")
	// Simulate a step that checks context
	select {
	case <-ctx.Done():
		return fmt.Errorf("serviceB: %w", ctx.Err())
	case <-time.After(50 * time.Millisecond):
		return serviceC(ctx)
	}
}

func serviceC(ctx context.Context) error {
	// Simulate a slow database call
	select {
	case <-ctx.Done():
		return fmt.Errorf("serviceC: %w", ctx.Err())
	case <-time.After(200 * time.Millisecond): // this will exceed our timeout
		fmt.Println("  serviceC: done")
		return nil
	}
}

func contextPropagation() {
	fmt.Println("\nContext propagation:")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := serviceA(ctx); err != nil {
		fmt.Printf("  chain failed: %v\n", err)
	}
}

func main() {
	backgroundAndTODO()
	withCancelDemo()
	withTimeoutDemo()
	withDeadlineDemo()
	withValueDemo()
	contextPropagation()
}
