// Package main demonstrates interfaces in Go.
// Topics: implicit implementation, composition, empty interface, Stringer, error.
package main

import (
	"fmt"
	"math"
)

// -----------------------------------------------------------------------
// SECTION 1: Interface Basics — Implicit Implementation
// -----------------------------------------------------------------------
// In Go, interfaces are satisfied IMPLICITLY.
// You don't write "implements Shape" anywhere.
// A type satisfies an interface simply by having the required methods.
// This decouples interface definitions from their implementations.

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct {
	Radius float64
}

// Circle implements Shape — just by having these two methods
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// printShape accepts any value that implements Shape
// It doesn't know or care if it's a Circle, Rectangle, or anything else
func printShape(s Shape) {
	fmt.Printf("  %T → area=%.2f perimeter=%.2f\n", s, s.Area(), s.Perimeter())
}

func interfaceBasics() {
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 4, Height: 6},
		Circle{Radius: 3},
	}

	fmt.Println("Shapes:")
	for _, s := range shapes {
		printShape(s)
	}

	// The interface variable holds both the type and value
	var s Shape = Circle{Radius: 1}
	fmt.Printf("  interface value: type=%T value=%v\n", s, s)
}

// -----------------------------------------------------------------------
// SECTION 2: Interface Composition
// -----------------------------------------------------------------------
// Interfaces can embed other interfaces to combine requirements.

type Reader interface {
	Read() string
}

type Writer interface {
	Write(data string)
}

// ReadWriter composes Reader and Writer
// A type must implement BOTH to satisfy ReadWriter
type ReadWriter interface {
	Reader
	Writer
}

// Buffer satisfies ReadWriter (has both Read and Write)
type Buffer struct {
	data string
}

func (b *Buffer) Read() string       { return b.data }
func (b *Buffer) Write(data string)  { b.data += data }

func interfaceComposition() {
	fmt.Println("\nInterface composition:")

	var rw ReadWriter = &Buffer{}
	rw.Write("hello ")
	rw.Write("world")
	fmt.Printf("  ReadWriter.Read(): %q\n", rw.Read())

	// A ReadWriter also satisfies Reader and Writer individually
	var r Reader = rw
	fmt.Printf("  Reader.Read():     %q\n", r.Read())
}

// -----------------------------------------------------------------------
// SECTION 3: The Stringer Interface
// -----------------------------------------------------------------------
// fmt.Stringer is a built-in interface:
//   type Stringer interface { String() string }
//
// Implement it to control how your type is printed by fmt functions.

type Temperature struct {
	Celsius float64
}

// By implementing String(), fmt.Println will call this automatically
func (t Temperature) String() string {
	return fmt.Sprintf("%.1f°C (%.1f°F)", t.Celsius, t.Celsius*9/5+32)
}

type Color struct {
	R, G, B uint8
}

func (c Color) String() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func stringer() {
	fmt.Println("\nStringer interface:")

	t := Temperature{Celsius: 100}
	fmt.Printf("  %v\n", t)    // calls t.String()
	fmt.Println(" ", t)         // also calls t.String()

	c := Color{255, 128, 0}
	fmt.Printf("  %v\n", c)
	fmt.Printf("  %s\n", c)
}

// -----------------------------------------------------------------------
// SECTION 4: The error Interface
// -----------------------------------------------------------------------
// error is a built-in interface:
//   type error interface { Error() string }
//
// Any type with an Error() string method is an error.
// Custom error types can carry extra context.

// ValidationError carries field name and message
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}

// NotFoundError carries what was not found
type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %d not found", e.Resource, e.ID)
}

func findUser(id int) error {
	if id <= 0 {
		return &ValidationError{Field: "id", Message: "must be positive"}
	}
	if id > 100 {
		return &NotFoundError{Resource: "user", ID: id}
	}
	return nil // nil means no error
}

func errorInterface() {
	fmt.Println("\nerror interface:")

	for _, id := range []int{0, 42, 999} {
		err := findUser(id)
		if err == nil {
			fmt.Printf("  findUser(%d): ok\n", id)
			continue
		}

		// Type-assert to get specific error details
		switch e := err.(type) {
		case *ValidationError:
			fmt.Printf("  findUser(%d): validation — field=%q msg=%q\n", id, e.Field, e.Message)
		case *NotFoundError:
			fmt.Printf("  findUser(%d): not found — %s#%d\n", id, e.Resource, e.ID)
		default:
			fmt.Printf("  findUser(%d): unknown error: %v\n", id, err)
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Empty Interface (any)
// -----------------------------------------------------------------------
// `any` is an alias for `interface{}` — satisfied by ALL types.
// It's Go's way of accepting "any value".
//
// Use it sparingly — you lose type safety. Prefer generics (Go 1.18+)
// or concrete interfaces when possible.

func printAnything(values ...any) {
	for _, v := range values {
		fmt.Printf("  %T: %v\n", v, v)
	}
}

func emptyInterface() {
	fmt.Println("\nEmpty interface (any):")

	printAnything(42, "hello", true, 3.14, []int{1, 2, 3}, nil)

	// Common use: heterogeneous container
	record := map[string]any{
		"name":  "Alice",
		"age":   30,
		"admin": true,
		"tags":  []string{"go", "backend"},
	}
	for k, v := range record {
		fmt.Printf("  record[%q] = %v (%T)\n", k, v, v)
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Interface Nil Pitfall
// -----------------------------------------------------------------------
// A nil interface is NOT the same as an interface holding a nil pointer.
// This is one of the most subtle bugs in Go.

type MyError struct{ msg string }
func (e *MyError) Error() string { return e.msg }

func mayFail(fail bool) error {
	// WRONG: this returns a non-nil interface (holding a nil *MyError)
	var err *MyError // nil pointer
	if fail {
		err = &MyError{"something went wrong"}
	}
	return err // DANGER: even when err is nil pointer, the interface is non-nil!
}

func mayFailCorrect(fail bool) error {
	// CORRECT: return the untyped nil directly
	if fail {
		return &MyError{"something went wrong"}
	}
	return nil // untyped nil → nil interface
}

func nilInterfacePitfall() {
	fmt.Println("\nInterface nil pitfall:")

	err := mayFail(false)
	fmt.Printf("  mayFail(false):        err=%v  isNil=%v  ← BUG!\n", err, err == nil)
	//                                                           false, even though *MyError is nil

	err2 := mayFailCorrect(false)
	fmt.Printf("  mayFailCorrect(false): err=%v  isNil=%v  ← correct\n", err2, err2 == nil)
}

func main() {
	interfaceBasics()
	interfaceComposition()
	stringer()
	errorInterface()
	emptyInterface()
	nilInterfacePitfall()
}
