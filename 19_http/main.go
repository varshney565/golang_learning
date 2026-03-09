// Package main demonstrates net/http — building HTTP servers and clients in Go.
// Topics: handlers, ServeMux, middleware, client, JSON API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: HTTP Handler Basics
// -----------------------------------------------------------------------
// An HTTP handler in Go is anything that implements http.Handler:
//   type Handler interface { ServeHTTP(ResponseWriter, *Request) }
//
// http.HandlerFunc is a function type that implements Handler,
// so any function with the right signature is a handler.

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// ResponseWriter writes the HTTP response
	// *Request contains everything about the incoming request
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK) // optional — 200 is the default
	fmt.Fprintln(w, "Hello, World!")
}

// -----------------------------------------------------------------------
// SECTION 2: JSON API Handler
// -----------------------------------------------------------------------

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type APIResponse struct {
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

// writeJSON is a helper to send JSON responses
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// usersHandler handles GET /users and POST /users
func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users := []User{
			{ID: 1, Name: "Alice", Age: 30},
			{ID: 2, Name: "Bob", Age: 25},
		}
		writeJSON(w, http.StatusOK, APIResponse{Data: users, Success: true})

	case http.MethodPost:
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{
				Error:   "invalid JSON: " + err.Error(),
				Success: false,
			})
			return
		}
		user.ID = 3 // simulate DB assignment
		writeJSON(w, http.StatusCreated, APIResponse{Data: user, Success: true})

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// -----------------------------------------------------------------------
// SECTION 3: Middleware
// -----------------------------------------------------------------------
// Middleware is a function that wraps a handler, adding behavior
// (logging, auth, rate limiting, etc.) without changing the handler.
// Pattern: func(http.Handler) http.Handler

// loggingMiddleware logs each request
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the actual handler
		next.ServeHTTP(w, r)

		fmt.Printf("  [LOG] %s %s — %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}

// authMiddleware checks for a Bearer token
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer secret-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return // stop the chain — don't call next
		}
		next.ServeHTTP(w, r)
	})
}

// chain applies multiple middleware in order: chain(h, m1, m2) → m1(m2(h))
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply in reverse so execution order matches argument order
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// -----------------------------------------------------------------------
// SECTION 4: ServeMux — Request Routing
// -----------------------------------------------------------------------

func buildMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, "Welcome to the Go API!")
	})

	mux.HandleFunc("/hello", helloHandler)

	// Protected route: wrap with auth middleware
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "secret data")
	})
	mux.Handle("/secret", chain(protected, loggingMiddleware, authMiddleware))

	mux.HandleFunc("/users", usersHandler)

	// Serve static files from a directory
	// mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return mux
}

// -----------------------------------------------------------------------
// SECTION 5: HTTP Client
// -----------------------------------------------------------------------

func httpClientDemo() {
	fmt.Println("\nHTTP Client:")

	// Always use a custom client with timeouts — NEVER use http.DefaultClient in production
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Use httptest to create a test server locally (no real network call)
	server := httptest.NewServer(http.HandlerFunc(usersHandler))
	defer server.Close()

	// GET request
	resp, err := client.Get(server.URL + "/users")
	if err != nil {
		fmt.Printf("  GET error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  GET /users: status=%d\n", resp.StatusCode)
	fmt.Printf("  body: %s", body)

	// POST request with JSON body
	newUser := User{Name: "Carol", Age: 28}
	payload, _ := json.Marshal(newUser)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/users", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")

	resp2, err := client.Do(req)
	if err != nil {
		fmt.Printf("  POST error: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("  POST /users: status=%d\n", resp2.StatusCode)
	fmt.Printf("  body: %s", body2)
}

// -----------------------------------------------------------------------
// SECTION 6: Request with Context (Cancellation + Timeout)
// -----------------------------------------------------------------------

func clientWithContext() {
	fmt.Println("\nClient with context:")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow handler
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintln(w, "slow response")
	}))
	defer server.Close()

	// Create request with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)

	client := &http.Client{}
	_, err := client.Do(req)
	if err != nil {
		fmt.Printf("  request cancelled: %v\n", err)
	}
}

// -----------------------------------------------------------------------
// SECTION 7: Testing Handlers with httptest
// -----------------------------------------------------------------------

func testHandlerDemo() {
	fmt.Println("\nhttptest (unit testing handlers):")

	// httptest.NewRecorder records the response without a real server
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	usersHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("  status: %d\n", resp.StatusCode)
	fmt.Printf("  content-type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("  body: %s", body)
}

func main() {
	fmt.Println("HTTP server/client demo (using httptest — no real server started):")

	httpClientDemo()
	clientWithContext()
	testHandlerDemo()

	// To actually start a server:
	// mux := buildMux()
	// log.Fatal(http.ListenAndServe(":8080", loggingMiddleware(mux)))
	fmt.Println("\nTo start a real server:")
	fmt.Println("  http.ListenAndServe(\":8080\", buildMux())")
}
