// Package main demonstrates graceful shutdown in Go.
// Topics: os.Signal, http.Server.Shutdown, cleanup on exit, shutdown timeout.
// Graceful shutdown is asked in nearly every backend Go interview.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Why Graceful Shutdown?
// -----------------------------------------------------------------------
// Without graceful shutdown:
//   - In-flight HTTP requests are abruptly terminated
//   - Database transactions may be left open
//   - Messages may be partially processed
//   - File writes may be incomplete
//
// Graceful shutdown:
//   1. Stop accepting NEW connections/work
//   2. Wait for IN-FLIGHT work to complete (with a timeout)
//   3. Release resources (close DB, flush logs, etc.)
//   4. Exit cleanly

// -----------------------------------------------------------------------
// SECTION 2: Signal Handling
// -----------------------------------------------------------------------
// SIGINT  → Ctrl+C (user interrupt)
// SIGTERM → sent by Kubernetes, systemd, Docker when stopping a container
// Both should trigger graceful shutdown.
//
// signal.NotifyContext is the modern way (Go 1.16+)

func signalHandling() {
	fmt.Println("Signal handling patterns:")

	// ── Pattern 1: signal.Notify (classic) ────────────────────
	quit := make(chan os.Signal, 1) // buffered: don't miss the signal
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Usage:
	// <-quit  // blocks until signal received
	// cleanup()
	signal.Stop(quit) // unsubscribe (cleanup for demo)

	// ── Pattern 2: signal.NotifyContext (Go 1.16+, preferred) ─
	// ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer stop()
	// <-ctx.Done()   // blocks until signal
	// cleanup()

	fmt.Println("  signal.Notify(quit, SIGINT, SIGTERM) — classic")
	fmt.Println("  signal.NotifyContext(ctx, SIGINT, SIGTERM) — modern")
}

// -----------------------------------------------------------------------
// SECTION 3: HTTP Server Graceful Shutdown
// -----------------------------------------------------------------------
// http.Server.Shutdown(ctx) — the built-in graceful shutdown:
//   1. Closes the listener (no new connections)
//   2. Closes idle connections
//   3. Waits for active connections to finish (up to ctx deadline)
//   4. Returns an error if ctx expires before all connections close

func runHTTPServer() {
	fmt.Println("\nHTTP Server graceful shutdown:")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow handler (in-flight request)
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, "hello")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Run server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		fmt.Println("  server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		fmt.Printf("  server error: %v\n", err)
	case sig := <-quit:
		fmt.Printf("  received signal: %v — shutting down...\n", sig)

		// Give in-flight requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("  shutdown error: %v\n", err)
		} else {
			fmt.Println("  server shut down gracefully")
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 4: Full Production Shutdown Pattern
// -----------------------------------------------------------------------
// Multiple components need to shut down in the right ORDER:
//   1. Stop accepting new work (HTTP, queue consumer, etc.)
//   2. Wait for in-flight work to drain
//   3. Close downstream dependencies (DB, cache, message queue)
//   4. Flush observability (metrics, traces, logs)

type App struct {
	httpServer *http.Server
	db         *fakeDB
	queue      *fakeQueue
	wg         sync.WaitGroup // tracks in-flight background work
}

// fakeDB simulates a database
type fakeDB struct{}
func (db *fakeDB) Close() error {
	fmt.Println("  [DB] connection pool closed")
	return nil
}

// fakeQueue simulates a message queue consumer
type fakeQueue struct {
	done chan struct{}
}

func (q *fakeQueue) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("  [Queue] consumer started")
		for {
			select {
			case <-q.done:
				fmt.Println("  [Queue] consumer stopped")
				return
			case <-time.After(100 * time.Millisecond):
				// process a message
			}
		}
	}()
}

func (q *fakeQueue) Stop() {
	close(q.done)
}

func (a *App) Start() {
	// Start HTTP server
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("  [HTTP] error: %v\n", err)
		}
	}()
	fmt.Println("  [HTTP] server started on :8081")

	// Start queue consumer
	a.queue.Start(&a.wg)
}

func (a *App) Shutdown() {
	fmt.Println("\n  beginning graceful shutdown...")

	// Step 1: Stop accepting new HTTP requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("  [HTTP] shutdown error: %v\n", err)
	} else {
		fmt.Println("  [HTTP] server stopped")
	}

	// Step 2: Stop queue consumer
	a.queue.Stop()

	// Step 3: Wait for all in-flight background work
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		fmt.Println("  all background work drained")
	case <-time.After(5 * time.Second):
		fmt.Println("  timeout waiting for background work")
	}

	// Step 4: Close resources
	a.db.Close()
	fmt.Println("  shutdown complete")
}

func fullShutdownPattern() {
	fmt.Println("\nFull production shutdown pattern:")

	app := &App{
		httpServer: &http.Server{
			Addr:    ":8081",
			Handler: http.DefaultServeMux,
		},
		db:    &fakeDB{},
		queue: &fakeQueue{done: make(chan struct{})},
	}

	app.Start()

	// Simulate running for a bit then getting a shutdown signal
	time.Sleep(150 * time.Millisecond)
	fmt.Println("  (simulating SIGTERM)")
	app.Shutdown()
}

// -----------------------------------------------------------------------
// SECTION 5: Shutdown with context propagation
// -----------------------------------------------------------------------
// Use a root context for the entire application.
// Cancel it on shutdown — all goroutines that respect context will stop.

func contextBasedShutdown() {
	fmt.Println("\nContext-based shutdown:")

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Worker that respects context
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("  worker %d: stopping (%v)\n", i, ctx.Err())
					return
				case <-time.After(50 * time.Millisecond):
					// do work
				}
			}
		}()
	}

	// Let workers run for a bit
	time.Sleep(120 * time.Millisecond)

	// Cancel the context — all workers stop
	fmt.Println("  cancelling root context...")
	cancel()

	// Wait with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("  all workers stopped cleanly")
	case <-time.After(2 * time.Second):
		fmt.Println("  shutdown timed out")
	}
}

func main() {
	signalHandling()
	// runHTTPServer()   // Uncomment to start a real server
	fullShutdownPattern()
	contextBasedShutdown()

	fmt.Println("\nGraceful shutdown checklist:")
	checklist := []string{
		"✓ Signal handler: SIGINT + SIGTERM",
		"✓ HTTP: server.Shutdown(ctx) with timeout",
		"✓ Background workers: done channel or context cancellation",
		"✓ WaitGroup: wait for in-flight work to drain",
		"✓ Resources: close DB, cache, queue connections in order",
		"✓ Timeout: overall shutdown deadline (e.g. 30s for Kubernetes)",
	}
	for _, item := range checklist {
		fmt.Printf("  %s\n", item)
	}
}
