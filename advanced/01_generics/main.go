// Package main demonstrates Generics (Go 1.18+).
// Topics: type parameters, constraints, generic functions, generic types.
// This is a KEY interview topic — generics are heavily asked about.
package main

import "fmt"

// -----------------------------------------------------------------------
// SECTION 1: Why Generics?
// -----------------------------------------------------------------------
// Before Go 1.18, you had to write duplicate code for every type,
// or use `any` (losing type safety), or code generation.
// Generics let you write type-safe code that works for multiple types.

// WITHOUT generics — duplicate code
func sumInts(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func sumFloat64s(nums []float64) float64 {
	total := 0.0
	for _, n := range nums {
		total += n
	}
	return total
}

// WITH generics — one function for all numeric types
// [T Number] is a TYPE PARAMETER with a CONSTRAINT
// Number is defined below — it restricts which types T can be

// -----------------------------------------------------------------------
// SECTION 2: Constraints
// -----------------------------------------------------------------------
// A constraint is an interface that restricts which types a type parameter can be.
// Use | to union types.  ~T means "any type whose underlying type is T".

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// Comparable is a built-in constraint — any type that supports ==
// any is also built-in — no constraints at all

// Custom ordered constraint (all types that support <, >, etc.)
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64 | ~string
}

// -----------------------------------------------------------------------
// SECTION 3: Generic Functions
// -----------------------------------------------------------------------

// Sum works for any type satisfying Number
func Sum[T Number](nums []T) T {
	var total T // zero value of T
	for _, n := range nums {
		total += n
	}
	return total
}

// Min returns the smaller of two Ordered values
func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Map applies fn to each element — like Python's map
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter keeps elements where fn returns true
func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce folds a slice into a single value
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// Contains checks if a slice contains a value
func Contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

func genericFunctions() {
	fmt.Println("Generic functions:")

	// Sum works for int, float64, etc.
	ints := []int{1, 2, 3, 4, 5}
	floats := []float64{1.1, 2.2, 3.3}
	fmt.Printf("  Sum(ints):   %d\n", Sum(ints))
	fmt.Printf("  Sum(floats): %.1f\n", Sum(floats))

	// Min with type inference
	fmt.Printf("  Min(3, 5) = %d\n", Min(3, 5))
	fmt.Printf("  Min(\"apple\", \"banana\") = %s\n", Min("apple", "banana"))

	// Map: []int → []string
	strs := Map(ints, func(n int) string { return fmt.Sprintf("item-%d", n) })
	fmt.Printf("  Map to strings: %v\n", strs)

	// Filter: keep evens
	evens := Filter(ints, func(n int) bool { return n%2 == 0 })
	fmt.Printf("  Filter evens: %v\n", evens)

	// Reduce: sum
	total := Reduce(ints, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("  Reduce sum: %d\n", total)

	// Contains
	fmt.Printf("  Contains(ints, 3): %v\n", Contains(ints, 3))
	fmt.Printf("  Contains(ints, 9): %v\n", Contains(ints, 9))
}

// -----------------------------------------------------------------------
// SECTION 4: Generic Types (Generic Data Structures)
// -----------------------------------------------------------------------

// Stack is a generic LIFO stack
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Size() int { return len(s.items) }

// Pair holds two values of potentially different types
type Pair[A, B any] struct {
	First  A
	Second B
}

func NewPair[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{First: a, Second: b}
}

// Optional represents a value that may or may not be present (like Rust's Option)
type Optional[T any] struct {
	value   T
	present bool
}

func Some[T any](v T) Optional[T]    { return Optional[T]{value: v, present: true} }
func None[T any]() Optional[T]       { return Optional[T]{} }
func (o Optional[T]) IsPresent() bool { return o.present }
func (o Optional[T]) Get() (T, bool) { return o.value, o.present }
func (o Optional[T]) OrElse(def T) T {
	if o.present {
		return o.value
	}
	return def
}

func genericTypes() {
	fmt.Println("\nGeneric types:")

	// Generic stack
	var intStack Stack[int]
	intStack.Push(1)
	intStack.Push(2)
	intStack.Push(3)
	for intStack.Size() > 0 {
		v, _ := intStack.Pop()
		fmt.Printf("  popped: %d\n", v)
	}

	// String stack
	var strStack Stack[string]
	strStack.Push("hello")
	strStack.Push("world")
	top, _ := strStack.Peek()
	fmt.Printf("  string stack top: %q\n", top)

	// Pair
	p := NewPair("user-123", 42)
	fmt.Printf("  pair: (%s, %d)\n", p.First, p.Second)

	// Optional
	found := Some(42)
	notFound := None[int]()
	fmt.Printf("  found.OrElse(0)    = %d\n", found.OrElse(0))
	fmt.Printf("  notFound.OrElse(0) = %d\n", notFound.OrElse(0))
}

// -----------------------------------------------------------------------
// SECTION 5: Multiple Type Parameters
// -----------------------------------------------------------------------

// MapKeys returns all keys of a map
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Zip combines two slices into a slice of pairs
func Zip[A, B any](as []A, bs []B) []Pair[A, B] {
	n := len(as)
	if len(bs) < n {
		n = len(bs)
	}
	result := make([]Pair[A, B], n)
	for i := 0; i < n; i++ {
		result[i] = NewPair(as[i], bs[i])
	}
	return result
}

func multipleTypeParams() {
	fmt.Println("\nMultiple type parameters:")

	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := MapKeys(m) // type inferred as []string
	fmt.Printf("  MapKeys: %v\n", keys)

	names := []string{"Alice", "Bob", "Carol"}
	ages := []int{30, 25, 35}
	pairs := Zip(names, ages)
	for _, p := range pairs {
		fmt.Printf("  %s: %d\n", p.First, p.Second)
	}
}

func main() {
	genericFunctions()
	genericTypes()
	multipleTypeParams()
}
