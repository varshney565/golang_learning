// Package main demonstrates http.RoundTripper — custom HTTP transport.
// Topics: RoundTripper interface, retry, logging, auth injection, circuit breaker.
// RoundTripper is a senior Go interview topic — it's how you customize HTTP clients.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: What is http.RoundTripper?
// -----------------------------------------------------------------------
// http.RoundTripper is the interface that actually sends HTTP requests:
//
//   type RoundTripper interface {
//       RoundTrip(*Request) (*Response, error)
//   }
//
// http.Client uses a RoundTripper to send requests.
// Default: http.DefaultTransport (manages keep-alive, connection pooling, TLS).
//
// You can wrap the default transport to add:
//   - Logging
//   - Authentication headers
//   - Retry logic
//   - Rate limiting
//   - Circuit breaking
//   - Metrics/tracing
//
// Rule: RoundTrip should NOT modify the request it receives.
//       Clone it first if you need to add headers.

// -----------------------------------------------------------------------
// SECTION 2: Logging Transport
// -----------------------------------------------------------------------

type LoggingTransport struct {
	Base http.RoundTripper // the wrapped transport
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Execute the actual request
	resp, err := t.base().RoundTrip(req)

	// Log after the request
	duration := time.Since(start)
	if err != nil {
		fmt.Printf("  [LOG] %s %s → ERROR %v (%v)\n", req.Method, req.URL, err, duration)
	} else {
		fmt.Printf("  [LOG] %s %s → %d (%v)\n", req.Method, req.URL, resp.StatusCode, duration)
	}

	return resp, err
}

func (t *LoggingTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// -----------------------------------------------------------------------
// SECTION 3: Auth Injection Transport
// -----------------------------------------------------------------------
// Automatically add Authorization header to every request.
// Much cleaner than setting headers on every individual request.

type AuthTransport struct {
	Token string
	Base  http.RoundTripper
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request — NEVER modify the original
	// (RoundTrip contract: don't mutate the request)
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.Token)

	return t.base().RoundTrip(clone)
}

func (t *AuthTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// -----------------------------------------------------------------------
// SECTION 4: Retry Transport
// -----------------------------------------------------------------------
// Automatically retry on transient errors (5xx, network timeouts).
// Important: only retry IDEMPOTENT requests (GET, HEAD, etc.) safely.
// For POST/PATCH, only retry if you know the operation is safe to repeat.

type RetryTransport struct {
	Base       http.RoundTripper
	MaxRetries int
	Backoff    time.Duration // wait between retries
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	maxRetries := t.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}
	backoff := t.Backoff
	if backoff == 0 {
		backoff = 100 * time.Millisecond
	}

	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms
			wait := backoff * time.Duration(1<<uint(attempt-1))
			fmt.Printf("  [RETRY] attempt %d after %v\n", attempt, wait)
			time.Sleep(wait)

			// If the body was consumed on last attempt, we can't retry POST etc.
			// For safety, only retry if there's no body or body is a GetBody func
			if req.Body != nil && req.GetBody == nil {
				break // can't retry — body was already consumed
			}
			if req.GetBody != nil {
				// Reset the body for retry
				req.Body, _ = req.GetBody()
			}
		}

		resp, err = t.base().RoundTrip(req)

		// Don't retry if no error or if error is context-related
		if err != nil {
			if req.Context().Err() != nil {
				return nil, err // context cancelled/timed out — stop retrying
			}
			continue // network error — retry
		}

		// Don't retry client errors (4xx) — they won't fix themselves
		if resp.StatusCode < 500 {
			return resp, nil
		}

		// 5xx — server error, retry
		// Must drain and close body before retry to reuse connection
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	return resp, err
}

func (t *RetryTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// -----------------------------------------------------------------------
// SECTION 5: Chaining Transports
// -----------------------------------------------------------------------
// Wrap transports like middleware: each adds a behavior.
// Order matters: outermost runs first.

func buildClient(token string) *http.Client {
	// From inside out: base → auth → retry → logging
	base := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	authed := &AuthTransport{
		Token: token,
		Base:  base,
	}

	retried := &RetryTransport{
		Base:       authed,
		MaxRetries: 3,
		Backoff:    50 * time.Millisecond,
	}

	logged := &LoggingTransport{Base: retried}

	return &http.Client{
		Transport: logged,
		Timeout:   10 * time.Second,
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Demo
// -----------------------------------------------------------------------

func demo() {
	fmt.Println("\nRoundTripper demo:")

	// Test server that tracks what it receives
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		auth := r.Header.Get("Authorization")
		fmt.Printf("  [SERVER] request %d, auth=%q\n", attemptCount, auth)

		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 — trigger retry
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "success")
	}))
	defer server.Close()

	client := buildClient("my-secret-token")

	resp, err := client.Get(server.URL + "/api/data")
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  final response: %d — %s", resp.StatusCode, body)
}

// -----------------------------------------------------------------------
// SECTION 7: Request Body with GetBody for Retries
// -----------------------------------------------------------------------
// To retry POST requests, set GetBody so the body can be re-read.

func retryablePost(client *http.Client, url string, bodyBytes []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Set GetBody so RetryTransport can reset the body on retry
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}

	return client.Do(req)
}

// -----------------------------------------------------------------------
// SECTION 8: Context Propagation in RoundTripper
// -----------------------------------------------------------------------
// Always pass request context through — never ignore it.

type TracingTransport struct {
	Base http.RoundTripper
}

func (t *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Extract trace ID from context (in real code, use opentelemetry)
	if traceID, ok := req.Context().Value("trace_id").(string); ok {
		req = req.Clone(req.Context())
		req.Header.Set("X-Trace-ID", traceID)
	}
	return t.base().RoundTrip(req)
}

func (t *TracingTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func contextDemo() {
	fmt.Println("\nContext in RoundTripper:")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		fmt.Printf("  [SERVER] X-Trace-ID: %q\n", traceID)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &TracingTransport{},
	}

	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, _ := client.Do(req)
	resp.Body.Close()
}

func main() {
	demo()
	contextDemo()

	fmt.Println("\nRoundTripper use cases:")
	uses := []string{
		"Logging: log every request/response",
		"Auth: inject Bearer token or API key on every request",
		"Retry: retry on 5xx or network errors",
		"Rate limiting: throttle outgoing requests",
		"Circuit breaker: stop calling a failing service",
		"Tracing: inject trace/span IDs",
		"Caching: return cached responses",
		"Metrics: record latency, error rates",
	}
	for _, u := range uses {
		fmt.Printf("  • %s\n", u)
	}
}
