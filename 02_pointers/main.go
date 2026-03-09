// Package main demonstrates pointers in Go.
// Topics: pointer basics, when to use pointers, nil pointers, pointer receivers.
package main

import "fmt"

// -----------------------------------------------------------------------
// SECTION 1: Pointer Basics
// -----------------------------------------------------------------------
// A pointer holds the memory ADDRESS of a value, not the value itself.
//
//   &x   → gives the address of x  (address-of operator)
//   *p   → gives the value at address p  (dereference operator)

func pointerBasics() {
	x := 42

	p := &x // p is *int — a pointer to int
	fmt.Printf("x = %d\n", x)
	fmt.Printf("address of x = %v\n", p)  // something like 0xc0000b4008
	fmt.Printf("value via p  = %d\n", *p) // dereference: same as x

	// Modifying via pointer changes the original variable
	*p = 100
	fmt.Printf("x after *p = 100 → x = %d\n", x)
}

// -----------------------------------------------------------------------
// SECTION 2: Why Pointers? — Pass by Reference
// -----------------------------------------------------------------------
// Go passes everything by VALUE. Without a pointer, the function gets
// its own copy and changes don't affect the caller.

// Does NOT modify the original — gets a copy
func doubleByValue(n int) {
	n = n * 2
	fmt.Printf("  inside doubleByValue: n = %d\n", n)
}

// DOES modify the original — receives the address
func doubleByPointer(n *int) {
	*n = *n * 2
	fmt.Printf("  inside doubleByPointer: *n = %d\n", *n)
}

func passByRef() {
	fmt.Println("\nPass by value vs pointer:")
	a := 5
	doubleByValue(a)
	fmt.Printf("  a after doubleByValue:   %d (unchanged)\n", a)

	doubleByPointer(&a)
	fmt.Printf("  a after doubleByPointer: %d (changed)\n", a)
}

// -----------------------------------------------------------------------
// SECTION 3: new() vs & (address-of)
// -----------------------------------------------------------------------
// Two ways to get a pointer to allocated memory:
//   new(T)   → allocates zero-value T, returns *T
//   &T{...}  → creates a T literal, returns *T (more common)

func newVsAddressOf() {
	fmt.Println("\nnew() vs &:")

	// new() — rarely used directly; gives a *T pointing to zero value
	p1 := new(int)
	fmt.Printf("  new(int): %v → value: %d\n", p1, *p1)

	// & on a composite literal — the idiomatic Go way
	type Point struct{ X, Y int }
	p2 := &Point{X: 3, Y: 4} // *Point
	fmt.Printf("  &Point{}: %v\n", p2)
}

// -----------------------------------------------------------------------
// SECTION 4: Nil Pointers — The Zero Value of a Pointer
// -----------------------------------------------------------------------
// The zero value of any pointer is nil. Dereferencing a nil pointer
// causes a runtime PANIC. Always check for nil before dereferencing.

func nilPointers() {
	fmt.Println("\nNil pointer safety:")

	var p *int // nil pointer
	fmt.Printf("  p = %v (nil? %v)\n", p, p == nil)

	// Safe pattern — always check nil before use
	if p != nil {
		fmt.Printf("  *p = %d\n", *p)
	} else {
		fmt.Println("  p is nil, skipping dereference")
	}

	// Assign and use
	n := 7
	p = &n
	fmt.Printf("  after p = &n: *p = %d\n", *p)
}

// -----------------------------------------------------------------------
// SECTION 5: Pointer Receivers on Methods
// -----------------------------------------------------------------------
// Methods can have either a value receiver or a pointer receiver.
//
//   Value receiver (v T)    → works on a COPY, cannot modify original
//   Pointer receiver (v *T) → works on the ORIGINAL, can modify it
//
// Use pointer receiver when:
//   1. The method needs to mutate the struct
//   2. The struct is large (avoid copying)

type Counter struct {
	count int
}

// Value receiver — does NOT modify the Counter
func (c Counter) Value() int {
	return c.count
}

// Pointer receiver — DOES modify the Counter
func (c *Counter) Increment() {
	c.count++
}

func (c *Counter) Reset() {
	c.count = 0
}

func pointerReceivers() {
	fmt.Println("\nPointer receivers:")

	c := Counter{}
	fmt.Printf("  initial: %d\n", c.Value())

	c.Increment()
	c.Increment()
	c.Increment()
	fmt.Printf("  after 3 increments: %d\n", c.Value())

	c.Reset()
	fmt.Printf("  after reset: %d\n", c.Value())

	// Go auto-takes address when calling pointer receiver on addressable value
	// c.Increment() is shorthand for (&c).Increment()
}

// -----------------------------------------------------------------------
// SECTION 6: Pointers to Structs
// -----------------------------------------------------------------------
// When you have a *Struct, Go lets you access fields directly with .
// instead of forcing you to write (*p).Field

type Person struct {
	Name string
	Age  int
}

func birthday(p *Person) {
	p.Age++ // same as (*p).Age++ — Go handles this automatically
}

func structPointers() {
	fmt.Println("\nStruct pointers:")

	person := Person{Name: "Alice", Age: 30}
	birthday(&person)
	fmt.Printf("  %s is now %d\n", person.Name, person.Age)
}

func main() {
	pointerBasics()
	passByRef()
	newVsAddressOf()
	nilPointers()
	pointerReceivers()
	structPointers()
}
