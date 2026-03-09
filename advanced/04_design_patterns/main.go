// Package main demonstrates common Go design patterns.
// These are heavily asked in interviews — functional options, builder, observer, etc.
package main

import (
	"fmt"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Functional Options Pattern
// -----------------------------------------------------------------------
// Problem: Constructors with many optional parameters become unwieldy.
// Solution: Each option is a function that configures the struct.
// This is THE idiomatic Go way to handle optional config.
//
// Used by: gRPC, zap logger, many popular libraries.

type Server struct {
	host    string
	port    int
	timeout time.Duration
	maxConn int
	tls     bool
}

// Option is a function type that configures a Server
type Option func(*Server)

// Each With* function returns an Option
func WithHost(host string) Option {
	return func(s *Server) { s.host = host }
}

func WithPort(port int) Option {
	return func(s *Server) { s.port = port }
}

func WithTimeout(d time.Duration) Option {
	return func(s *Server) { s.timeout = d }
}

func WithMaxConn(n int) Option {
	return func(s *Server) { s.maxConn = n }
}

func WithTLS() Option {
	return func(s *Server) { s.tls = true }
}

// NewServer applies all options on top of sensible defaults
func NewServer(opts ...Option) *Server {
	s := &Server{ // sensible defaults
		host:    "localhost",
		port:    8080,
		timeout: 30 * time.Second,
		maxConn: 100,
	}
	for _, opt := range opts {
		opt(s) // apply each option
	}
	return s
}

func functionalOptions() {
	fmt.Println("Functional Options pattern:")

	// Use only the options you need
	s1 := NewServer() // all defaults
	s2 := NewServer(
		WithHost("0.0.0.0"),
		WithPort(9090),
		WithTLS(),
		WithTimeout(10*time.Second),
	)

	fmt.Printf("  s1: host=%s port=%d tls=%v\n", s1.host, s1.port, s1.tls)
	fmt.Printf("  s2: host=%s port=%d tls=%v\n", s2.host, s2.port, s2.tls)
}

// -----------------------------------------------------------------------
// SECTION 2: Builder Pattern
// -----------------------------------------------------------------------
// Similar to functional options but uses method chaining.
// Good for SQL-like query builders, complex object construction.

type QueryBuilder struct {
	table      string
	conditions []string
	orderBy    string
	limit      int
}

func NewQuery(table string) *QueryBuilder {
	return &QueryBuilder{table: table, limit: -1}
}

func (q *QueryBuilder) Where(condition string) *QueryBuilder {
	q.conditions = append(q.conditions, condition)
	return q // return self for chaining
}

func (q *QueryBuilder) OrderBy(col string) *QueryBuilder {
	q.orderBy = col
	return q
}

func (q *QueryBuilder) Limit(n int) *QueryBuilder {
	q.limit = n
	return q
}

func (q *QueryBuilder) Build() string {
	sql := fmt.Sprintf("SELECT * FROM %s", q.table)
	for i, c := range q.conditions {
		if i == 0 {
			sql += " WHERE " + c
		} else {
			sql += " AND " + c
		}
	}
	if q.orderBy != "" {
		sql += " ORDER BY " + q.orderBy
	}
	if q.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	return sql
}

func builderPattern() {
	fmt.Println("\nBuilder pattern:")

	query := NewQuery("users").
		Where("age > 18").
		Where("active = true").
		OrderBy("name").
		Limit(10).
		Build()

	fmt.Printf("  %s\n", query)
}

// -----------------------------------------------------------------------
// SECTION 3: Observer / Event Bus Pattern
// -----------------------------------------------------------------------
// Decouple event producers from consumers.
// Producers publish events; consumers subscribe to event types.

type EventType string

const (
	UserCreated EventType = "user.created"
	UserDeleted EventType = "user.deleted"
)

type Event struct {
	Type    EventType
	Payload any
}

type Handler func(Event)

type EventBus struct {
	subscribers map[EventType][]Handler
}

func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[EventType][]Handler)}
}

func (eb *EventBus) Subscribe(eventType EventType, handler Handler) {
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
	for _, handler := range eb.subscribers[event.Type] {
		handler(event) // in production, run in goroutines for async
	}
}

func observerPattern() {
	fmt.Println("\nObserver / Event Bus pattern:")

	bus := NewEventBus()

	// Register handlers
	bus.Subscribe(UserCreated, func(e Event) {
		fmt.Printf("  [Email Service] Send welcome email to %v\n", e.Payload)
	})
	bus.Subscribe(UserCreated, func(e Event) {
		fmt.Printf("  [Audit Log] User created: %v\n", e.Payload)
	})
	bus.Subscribe(UserDeleted, func(e Event) {
		fmt.Printf("  [Cleanup] Remove data for user: %v\n", e.Payload)
	})

	// Publish events
	bus.Publish(Event{Type: UserCreated, Payload: "alice@example.com"})
	bus.Publish(Event{Type: UserDeleted, Payload: "bob@example.com"})
}

// -----------------------------------------------------------------------
// SECTION 4: Middleware Chain (Pipeline of Functions)
// -----------------------------------------------------------------------
// Compose functions where each wraps the next.
// Used in HTTP middleware, request processing, etc.

type Middleware func(string) string

func chain(middlewares ...Middleware) Middleware {
	return func(input string) string {
		result := input
		for _, m := range middlewares {
			result = m(result)
		}
		return result
	}
}

func middlewareChain() {
	fmt.Println("\nMiddleware chain:")

	upper := func(s string) string { return "[UPPER:" + s + "]" }
	prefix := func(s string) string { return "PREFIX-" + s }
	trim := func(s string) string { return s + "-TRIMMED" }

	pipeline := chain(upper, prefix, trim)
	result := pipeline("hello")
	fmt.Printf("  input: \"hello\"\n")
	fmt.Printf("  result: %s\n", result)
}

// -----------------------------------------------------------------------
// SECTION 5: Singleton Pattern
// -----------------------------------------------------------------------
// Already shown in phase3/04_sync with sync.Once.
// Reminder: use sync.Once for thread-safe lazy initialization.

// -----------------------------------------------------------------------
// SECTION 6: Strategy Pattern
// -----------------------------------------------------------------------
// Define a family of algorithms, encapsulate each one, make them interchangeable.
// In Go: just use interfaces or function values.

type SortStrategy func([]int) []int

type Sorter struct {
	strategy SortStrategy
}

func (s *Sorter) Sort(data []int) []int {
	return s.strategy(data)
}

func bubbleSort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

func reverseSort(data []int) []int {
	result := bubbleSort(data)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func strategyPattern() {
	fmt.Println("\nStrategy pattern:")

	data := []int{5, 2, 8, 1, 9, 3}
	sorter := &Sorter{}

	sorter.strategy = bubbleSort
	fmt.Printf("  bubble sort: %v\n", sorter.Sort(data))

	sorter.strategy = reverseSort
	fmt.Printf("  reverse sort: %v\n", sorter.Sort(data))
}

// -----------------------------------------------------------------------
// SECTION 7: Decorator Pattern
// -----------------------------------------------------------------------
// Add behavior to an object without modifying it.
// In Go: wrap an interface with another type that embeds it.

type Storage interface {
	Save(key, value string) error
	Load(key string) (string, error)
}

// In-memory storage
type MemStorage struct {
	data map[string]string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{data: make(map[string]string)}
}
func (m *MemStorage) Save(k, v string) error { m.data[k] = v; return nil }
func (m *MemStorage) Load(k string) (string, error) {
	v, ok := m.data[k]
	if !ok {
		return "", fmt.Errorf("key %q not found", k)
	}
	return v, nil
}

// LoggingStorage wraps any Storage and adds logging
type LoggingStorage struct {
	wrapped Storage // embedded behavior
}

func (l *LoggingStorage) Save(k, v string) error {
	fmt.Printf("  [LOG] Save(%q, %q)\n", k, v)
	return l.wrapped.Save(k, v)
}

func (l *LoggingStorage) Load(k string) (string, error) {
	v, err := l.wrapped.Load(k)
	fmt.Printf("  [LOG] Load(%q) = %q, err=%v\n", k, v, err)
	return v, err
}

func decoratorPattern() {
	fmt.Println("\nDecorator pattern:")

	base := NewMemStorage()
	logged := &LoggingStorage{wrapped: base}

	// Use logged exactly like Storage
	logged.Save("name", "Alice")
	logged.Load("name")
	logged.Load("missing")
}

func main() {
	functionalOptions()
	builderPattern()
	observerPattern()
	middlewareChain()
	strategyPattern()
	decoratorPattern()
}
