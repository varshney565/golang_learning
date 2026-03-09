// Package main demonstrates Go's type system at an intermediate level.
// Topics: basic types, zero values, type aliases, type assertions, type switches.
package main

import "fmt"

// -----------------------------------------------------------------------
// SECTION 1: Basic Types & Zero Values
// -----------------------------------------------------------------------
// Every type in Go has a "zero value" — the default when you declare a
// variable without initializing it. This avoids undefined behavior.

func zeroValues() {
	var i int       // 0
	var f float64   // 0.0
	var b bool      // false
	var s string    // "" (empty string)
	var p *int      // nil (pointer)

	fmt.Println("Zero values:")
	fmt.Printf("  int:     %v\n", i)
	fmt.Printf("  float64: %v\n", f)
	fmt.Printf("  bool:    %v\n", b)
	fmt.Printf("  string:  %q\n", s)
	fmt.Printf("  *int:    %v\n", p)
}

// -----------------------------------------------------------------------
// SECTION 2: Type Aliases vs Type Definitions
// -----------------------------------------------------------------------
// A type DEFINITION creates a brand new type. Even if the underlying type
// is the same, Go treats them as different — you can't mix them without
// explicit conversion.
//
// A type ALIAS (=) is just another name for the same type.

type Celsius float64    // new type — not interchangeable with float64 directly
type Fahrenheit float64 // new type

type MyString = string // alias — exactly the same as string

// Method on a custom type — you can only add methods to types you define
// in the same package.
func (c Celsius) ToFahrenheit() Fahrenheit {
	return Fahrenheit(c*9/5 + 32)
}

func typeDefinitions() {
	var temp Celsius = 100
	fmt.Printf("\nType definitions:\n")
	fmt.Printf("  %v°C = %v°F\n", temp, temp.ToFahrenheit())

	// This would be a compile error:
	// var raw float64 = temp  ← cannot use Celsius as float64

	// Explicit conversion is required:
	raw := float64(temp)
	fmt.Printf("  As float64: %v\n", raw)

	// Alias behaves exactly like the original type
	var ms MyString = "hello" // works exactly like string
	fmt.Printf("  MyString: %q\n", ms)
}

// -----------------------------------------------------------------------
// SECTION 3: Type Assertions
// -----------------------------------------------------------------------
// When you have a value stored in an interface (like `any`/`interface{}`),
// you can "assert" what the underlying concrete type is.
// Two forms:
//   value := i.(Type)           → panics if wrong type
//   value, ok := i.(Type)       → safe, ok is false if wrong type

func typeAssertions() {
	fmt.Println("\nType assertions:")

	var i any = "hello world" // `any` is an alias for `interface{}`

	// Safe assertion — always prefer this form
	s, ok := i.(string)
	fmt.Printf("  i.(string) → value=%q, ok=%v\n", s, ok)

	n, ok := i.(int)
	fmt.Printf("  i.(int)    → value=%v, ok=%v\n", n, ok)

	// Unsafe assertion — panics if the type is wrong
	// s2 := i.(int)  ← this would panic at runtime
}

// -----------------------------------------------------------------------
// SECTION 4: Type Switches
// -----------------------------------------------------------------------
// A type switch is like a regular switch but branches on the dynamic type
// of an interface value. Very common when handling `any` values.

func describe(i any) string {
	switch v := i.(type) {
	// v is automatically typed correctly in each case
	case int:
		return fmt.Sprintf("int: %d (doubled: %d)", v, v*2)
	case string:
		return fmt.Sprintf("string: %q (length: %d)", v, len(v))
	case bool:
		return fmt.Sprintf("bool: %v", v)
	case []int:
		return fmt.Sprintf("[]int with %d elements", len(v))
	case nil:
		return "nil value"
	default:
		// %T prints the type name
		return fmt.Sprintf("unknown type: %T", v)
	}
}

func typeSwitches() {
	fmt.Println("\nType switches:")
	values := []any{42, "gopher", true, []int{1, 2, 3}, nil, 3.14}
	for _, v := range values {
		fmt.Printf("  %v\n", describe(v))
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Constants and iota
// -----------------------------------------------------------------------
// `iota` is a special constant that auto-increments within a const block.
// It resets to 0 at the start of each const block.

type Direction int

const (
	North Direction = iota // 0
	East                   // 1
	South                  // 2
	West                   // 3
)

func (d Direction) String() string {
	return [...]string{"North", "East", "South", "West"}[d]
}

type ByteSize float64

const (
	_           = iota             // blank identifier — ignore first value (0)
	KB ByteSize = 1 << (10 * iota) // 1 << 10 = 1024
	MB                             // 1 << 20
	GB                             // 1 << 30
)

func constants() {
	fmt.Println("\nConstants with iota:")
	fmt.Printf("  North=%v East=%v South=%v West=%v\n", North, East, South, West)
	fmt.Printf("  KB=%.0f MB=%.0f GB=%.0f\n", float64(KB), float64(MB), float64(GB))
}

func main() {
	zeroValues()
	typeDefinitions()
	typeAssertions()
	typeSwitches()
	constants()
}
