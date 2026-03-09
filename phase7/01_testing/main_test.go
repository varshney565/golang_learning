// Package main demonstrates Go testing patterns.
// Topics: table-driven tests, subtests, benchmarks, mocks, httptest.
//
// Run tests:
//   go test ./...
//   go test -v ./...          (verbose)
//   go test -run TestAdd ./.. (run specific test)
//   go test -bench=. ./...    (run benchmarks)
//   go test -race ./...       (race detector)
//   go test -cover ./...      (coverage)

package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------
// Functions under test (normally in a separate .go file)
// -----------------------------------------------------------------------

func Add(a, b int) int { return a + b }

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// -----------------------------------------------------------------------
// SECTION 1: Simple Test
// -----------------------------------------------------------------------
// Test functions must start with Test, take *testing.T, return nothing.

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
		// t.Fatalf → like Errorf but stops the test immediately
		// t.Logf   → print info (only shown with -v or on failure)
	}
}

// -----------------------------------------------------------------------
// SECTION 2: Table-Driven Tests — The Go Idiom
// -----------------------------------------------------------------------
// Run many input/output pairs through the same test logic.
// Easier to maintain, easy to add cases.

func TestAddTable(t *testing.T) {
	tests := []struct {
		name string // descriptive test case name
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"negative", -1, -1, -2},
		{"zero", 0, 0, 0},
		{"mixed", -5, 3, -2},
		{"large", 1000000, 2000000, 3000000},
	}

	for _, tt := range tests {
		// t.Run creates a subtest — run individually with: go test -run TestAddTable/positive
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDivideTable(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool   // true if we expect an error
	}{
		{"normal", 10, 2, 5, false},
		{"decimal", 7, 2, 3.5, false},
		{"divide by zero", 5, 0, 0, true},
		{"negative", -6, 3, -2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Divide(%v, %v) error = %v, wantErr = %v", tt.a, tt.b, err, tt.wantErr)
				return
			}

			// Check result only when no error expected
			if !tt.wantErr && got != tt.want {
				t.Errorf("Divide(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// SECTION 3: Testing with Interfaces (Mocking)
// -----------------------------------------------------------------------
// Define an interface for external dependencies, inject a mock in tests.

type UserStore interface {
	GetUser(id int) (string, error)
}

type UserService struct {
	store UserStore
}

func (s *UserService) Greet(id int) (string, error) {
	name, err := s.store.GetUser(id)
	if err != nil {
		return "", fmt.Errorf("greet: %w", err)
	}
	return "Hello, " + name + "!", nil
}

// Mock implementation — used in tests only
type mockUserStore struct {
	users map[int]string
	err   error // if non-nil, all calls return this error
}

func (m *mockUserStore) GetUser(id int) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	name, ok := m.users[id]
	if !ok {
		return "", errors.New("not found")
	}
	return name, nil
}

func TestUserServiceGreet(t *testing.T) {
	tests := []struct {
		name    string
		store   UserStore
		id      int
		want    string
		wantErr bool
	}{
		{
			name:  "existing user",
			store: &mockUserStore{users: map[int]string{1: "Alice"}},
			id:    1,
			want:  "Hello, Alice!",
		},
		{
			name:    "missing user",
			store:   &mockUserStore{users: map[int]string{}},
			id:      99,
			wantErr: true,
		},
		{
			name:    "store error",
			store:   &mockUserStore{err: errors.New("db down")},
			id:      1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &UserService{store: tt.store}
			got, err := svc.Greet(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Greet(%d) error = %v, wantErr = %v", tt.id, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Greet(%d) = %q; want %q", tt.id, got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// SECTION 4: HTTP Handler Tests with httptest
// -----------------------------------------------------------------------

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "pong")
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatusCode int
		wantBody       string
	}{
		{"GET returns pong", http.MethodGet, http.StatusOK, "pong"},
		{"POST not allowed", http.MethodPost, http.StatusMethodNotAllowed, "method not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fake request
			req := httptest.NewRequest(tt.method, "/ping", nil)
			// Create a response recorder
			w := httptest.NewRecorder()

			pingHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("status = %d; want %d", resp.StatusCode, tt.wantStatusCode)
			}

			body := strings.TrimSpace(w.Body.String())
			if !strings.Contains(body, tt.wantBody) {
				t.Errorf("body = %q; want to contain %q", body, tt.wantBody)
			}
		})
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Benchmarks
// -----------------------------------------------------------------------
// Benchmark functions start with Benchmark, take *testing.B.
// Run with: go test -bench=. -benchmem ./...
//
// b.N is automatically adjusted by the testing framework to run enough
// iterations for a stable measurement.

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(42, 58) // code under benchmark
	}
}

func BenchmarkReverse(b *testing.B) {
	s := "Hello, World! This is a longer string for benchmarking."
	b.ResetTimer() // reset timer after setup
	for i := 0; i < b.N; i++ {
		Reverse(s)
	}
}

// Compare two implementations
func reverseV2(s string) string {
	n := len(s)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = s[n-1-i]
	}
	return string(b)
}

func BenchmarkReverseV2(b *testing.B) {
	s := "Hello, World! This is a longer string for benchmarking."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reverseV2(s)
	}
}

// -----------------------------------------------------------------------
// SECTION 6: TestMain — Setup and Teardown for the Entire Package
// -----------------------------------------------------------------------
// TestMain runs before any tests. Use for:
//   - Starting/stopping test databases
//   - Setting up shared fixtures
//   - Custom flags

func TestMain(m *testing.M) {
	// Setup — runs before all tests
	fmt.Println("  [TestMain] setting up test suite")

	// m.Run() runs all the tests
	// exitCode := m.Run()

	// Teardown — runs after all tests
	fmt.Println("  [TestMain] tearing down test suite")

	// os.Exit(exitCode) — exit with the test result code
	// For this demo, just run normally
	m.Run()
}
