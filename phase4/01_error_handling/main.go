// Package main demonstrates idiomatic error handling in Go.
// Topics: error wrapping, errors.Is, errors.As, sentinel errors, custom errors.
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// -----------------------------------------------------------------------
// SECTION 1: Sentinel Errors
// -----------------------------------------------------------------------
// A sentinel error is a package-level var that represents a specific condition.
// Callers compare with errors.Is() to check for it.
// Convention: name starts with Err.

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
)

func findUser(id int) (string, error) {
	switch id {
	case 1:
		return "Alice", nil
	case 2:
		return "", ErrUnauthorized
	default:
		return "", ErrNotFound
	}
}

func sentinelErrors() {
	fmt.Println("Sentinel errors:")

	for _, id := range []int{1, 2, 3} {
		user, err := findUser(id)
		if err == nil {
			fmt.Printf("  id=%d: user=%s\n", id, user)
			continue
		}

		// Use errors.Is — not direct == (handles wrapping)
		switch {
		case errors.Is(err, ErrNotFound):
			fmt.Printf("  id=%d: not found\n", id)
		case errors.Is(err, ErrUnauthorized):
			fmt.Printf("  id=%d: access denied\n", id)
		default:
			fmt.Printf("  id=%d: %v\n", id, err)
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 2: Custom Error Types
// -----------------------------------------------------------------------
// When you need to attach extra context (field name, HTTP status, etc.),
// create a struct that implements the error interface.

type ValidationError struct {
	Field   string
	Value   any
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: field=%q value=%v — %s", e.Field, e.Value, e.Message)
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func parseAge(s string) (int, error) {
	age, err := strconv.Atoi(s)
	if err != nil {
		return 0, &ValidationError{
			Field:   "age",
			Value:   s,
			Message: "must be a number",
		}
	}
	if age < 0 || age > 150 {
		return 0, &ValidationError{
			Field:   "age",
			Value:   age,
			Message: "must be between 0 and 150",
		}
	}
	return age, nil
}

func customErrors() {
	fmt.Println("\nCustom error types:")

	for _, input := range []string{"25", "abc", "-5", "200"} {
		age, err := parseAge(input)
		if err == nil {
			fmt.Printf("  parseAge(%q) = %d\n", input, age)
			continue
		}

		// errors.As extracts the concrete type if it matches
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Printf("  parseAge(%q): field=%s msg=%s\n", input, ve.Field, ve.Message)
		} else {
			fmt.Printf("  parseAge(%q): %v\n", input, err)
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 3: Error Wrapping — fmt.Errorf with %w
// -----------------------------------------------------------------------
// Use fmt.Errorf("...: %w", err) to wrap an error with additional context.
// The %w verb wraps the error, making it detectable via errors.Is/As.
// Creates an error chain: outer → inner → ... → root

func readConfig(path string) error {
	return fmt.Errorf("readConfig: file open failed for %q: %w", path, ErrNotFound)
}

func loadApp(configPath string) error {
	if err := readConfig(configPath); err != nil {
		return fmt.Errorf("loadApp: %w", err) // wrap again
	}
	return nil
}

func errorWrapping() {
	fmt.Println("\nError wrapping:")

	err := loadApp("/etc/app.conf")
	if err != nil {
		fmt.Printf("  error: %v\n", err)

		// errors.Is unwraps the chain to find ErrNotFound
		fmt.Printf("  is ErrNotFound: %v\n", errors.Is(err, ErrNotFound))

		// errors.Unwrap gets one level up
		inner := errors.Unwrap(err)
		fmt.Printf("  unwrapped once: %v\n", inner)
	}
}

// -----------------------------------------------------------------------
// SECTION 4: errors.Is vs errors.As
// -----------------------------------------------------------------------
// errors.Is(err, target) → checks if ANY error in the chain == target
//                          (uses == comparison, or custom Is() method)
// errors.As(err, &target) → checks if ANY error in the chain is of type T
//                           (assigns to target if found)

func isVsAs() {
	fmt.Println("\nerrors.Is vs errors.As:")

	// Wrap a ValidationError
	original := &ValidationError{Field: "email", Value: "bad", Message: "invalid format"}
	wrapped := fmt.Errorf("createUser failed: %w", original)

	// errors.Is compares by VALUE — won't work for custom types unless Is() is implemented
	// Sentinel errors work because they're pointer-equal
	fmt.Printf("  Is(ErrNotFound):    %v\n", errors.Is(wrapped, ErrNotFound))

	// errors.As checks the TYPE — works for custom error types
	var ve *ValidationError
	fmt.Printf("  As(ValidationError): %v\n", errors.As(wrapped, &ve))
	if ve != nil {
		fmt.Printf("  extracted field: %s\n", ve.Field)
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Implementing errors.Is on Custom Types
// -----------------------------------------------------------------------
// For sentinel-like custom errors, implement Is() to control equality.

type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %d not found", e.Resource, e.ID)
}

// Is makes errors.Is work based on resource type (ignore the specific ID)
func (e *NotFoundError) Is(target error) bool {
	t, ok := target.(*NotFoundError)
	if !ok {
		return false
	}
	// Match if resource matches (or target has empty resource = wildcard)
	return t.Resource == "" || e.Resource == t.Resource
}

func customIs() {
	fmt.Println("\nCustom Is() method:")

	err := fmt.Errorf("lookup failed: %w", &NotFoundError{Resource: "user", ID: 42})

	// Match any NotFoundError
	anyNotFound := &NotFoundError{}
	fmt.Printf("  Is(any NotFoundError):  %v\n", errors.Is(err, anyNotFound))

	// Match specific resource
	userNotFound := &NotFoundError{Resource: "user"}
	fmt.Printf("  Is(user NotFoundError): %v\n", errors.Is(err, userNotFound))

	// Won't match wrong resource
	postNotFound := &NotFoundError{Resource: "post"}
	fmt.Printf("  Is(post NotFoundError): %v\n", errors.Is(err, postNotFound))
}

// -----------------------------------------------------------------------
// SECTION 6: Error Handling Patterns
// -----------------------------------------------------------------------

// Pattern: early return to reduce nesting
func processData(input string) error {
	if input == "" {
		return fmt.Errorf("processData: %w", ErrInvalidInput)
	}

	value, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("processData: parse failed: %w", ErrInvalidInput)
	}

	if value < 0 {
		return fmt.Errorf("processData: %w", &ValidationError{
			Field: "value", Value: value, Message: "must be non-negative",
		})
	}

	fmt.Printf("  processed: %d\n", value)
	return nil
}

// Pattern: aggregate multiple errors
type MultiError struct {
	Errors []error
}

func (m *MultiError) Error() string {
	s := fmt.Sprintf("%d errors occurred:\n", len(m.Errors))
	for i, err := range m.Errors {
		s += fmt.Sprintf("  [%d] %v\n", i+1, err)
	}
	return s
}

func (m *MultiError) Add(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

func (m *MultiError) OrNil() error {
	if len(m.Errors) == 0 {
		return nil
	}
	return m
}

func errorPatterns() {
	fmt.Println("\nError handling patterns:")

	// Early return
	for _, input := range []string{"42", "", "abc"} {
		if err := processData(input); err != nil {
			fmt.Printf("  processData(%q): %v\n", input, err)
		}
	}

	// Multi-error aggregation
	var errs MultiError
	errs.Add(fmt.Errorf("step 1 failed"))
	errs.Add(nil) // nil is ignored
	errs.Add(fmt.Errorf("step 3 failed"))

	if err := errs.OrNil(); err != nil {
		fmt.Printf("\n  %v", err)
	}
}

func main() {
	sentinelErrors()
	customErrors()
	errorWrapping()
	isVsAs()
	customIs()
	errorPatterns()
}
