// Package main demonstrates maps in Go.
// Topics: creation, CRUD, nil maps, iteration order, existence checks, delete.
package main

import (
	"fmt"
	"sort"
)

// -----------------------------------------------------------------------
// SECTION 1: Creating Maps
// -----------------------------------------------------------------------
// A map is Go's built-in hash map (dictionary / hash table).
// Syntax: map[KeyType]ValueType
//
// Key type must be comparable (==) — slices, maps, funcs cannot be keys.
// Common key types: string, int, structs with comparable fields.

func creatingMaps() {
	// Map literal — initialized and ready to use
	ages := map[string]int{
		"Alice": 30,
		"Bob":   25,
		"Carol": 35,
	}
	fmt.Printf("literal:  %v\n", ages)

	// make — creates an empty map ready for use
	// Optional second argument is initial capacity hint (not a limit)
	scores := make(map[string]int)
	scores["Alice"] = 95
	scores["Bob"] = 87
	fmt.Printf("make:     %v\n", scores)

	// Var declaration — nil map (DO NOT write to it yet!)
	var m map[string]int
	fmt.Printf("var decl: %v  nil=%v\n", m, m == nil)
}

// -----------------------------------------------------------------------
// SECTION 2: Nil Map — The Most Common Gotcha
// -----------------------------------------------------------------------
// Reading from a nil map is safe — returns zero value.
// Writing to a nil map causes a PANIC.
// Always initialize a map before writing.

func nilMapGotcha() {
	fmt.Println("\nNil map behavior:")

	var m map[string]int // nil map

	// Reading is safe — returns zero value
	val := m["missing"]
	fmt.Printf("  reading nil map: %d (zero value, no panic)\n", val)

	// Writing to nil map panics — uncomment to see:
	// m["key"] = 1  // panic: assignment to entry in nil map

	// Fix: initialize first
	m = make(map[string]int)
	m["key"] = 1
	fmt.Printf("  after init: %v\n", m)
}

// -----------------------------------------------------------------------
// SECTION 3: The "comma ok" Pattern — Checking Key Existence
// -----------------------------------------------------------------------
// When you read a map, you always get a value back (zero if missing).
// To distinguish "key present with zero value" from "key not present",
// use the two-value form: value, ok := m[key]

func keyExistence() {
	fmt.Println("\nKey existence check:")

	config := map[string]string{
		"host": "localhost",
		"port": "8080",
	}

	// One-value form — can't tell if key exists or value is zero
	timeout := config["timeout"]
	fmt.Printf("  config[\"timeout\"] = %q (empty string — but does key exist?)\n", timeout)

	// Two-value form — ok is true only if the key was present
	host, ok := config["host"]
	fmt.Printf("  host:    %q  ok=%v\n", host, ok)

	timeout2, ok2 := config["timeout"]
	fmt.Printf("  timeout: %q  ok=%v (not set)\n", timeout2, ok2)

	// Common pattern: check then use
	if db, ok := config["db"]; ok {
		fmt.Printf("  db: %s\n", db)
	} else {
		fmt.Println("  db key not found, using default")
	}
}

// -----------------------------------------------------------------------
// SECTION 4: Deleting Keys
// -----------------------------------------------------------------------

func deletingKeys() {
	fmt.Println("\nDeleting keys:")

	m := map[string]int{"a": 1, "b": 2, "c": 3}
	fmt.Printf("  before: %v\n", m)

	delete(m, "b")
	fmt.Printf("  after delete(m, \"b\"): %v\n", m)

	// Deleting a non-existent key is a no-op (no panic)
	delete(m, "xyz")
	fmt.Printf("  after delete(m, \"xyz\"): %v (unchanged)\n", m)
}

// -----------------------------------------------------------------------
// SECTION 5: Iteration — Order is NOT Guaranteed
// -----------------------------------------------------------------------
// Map iteration in Go is deliberately randomized.
// If you need sorted output, sort the keys first.

func iteration() {
	fmt.Println("\nIteration:")

	m := map[string]int{"banana": 3, "apple": 1, "cherry": 2}

	// Unordered — different order each run
	fmt.Println("  unordered iteration:")
	for k, v := range m {
		fmt.Printf("    %s: %d\n", k, v)
	}

	// Sorted iteration — extract keys, sort, then iterate
	fmt.Println("  sorted iteration:")
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("    %s: %d\n", k, m[k])
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Maps as Sets
// -----------------------------------------------------------------------
// Go has no built-in set type. Use map[T]struct{} as a set.
// struct{} takes zero bytes of memory — more efficient than map[T]bool.

func mapsAsSets() {
	fmt.Println("\nMap as set:")

	// Build a set of unique words
	words := []string{"go", "is", "go", "fast", "is", "go"}
	seen := make(map[string]struct{})

	for _, w := range words {
		seen[w] = struct{}{} // struct{}{} is the zero-size value
	}

	fmt.Printf("  unique words: ")
	for w := range seen {
		fmt.Printf("%s ", w)
	}
	fmt.Println()

	// Check membership
	if _, ok := seen["go"]; ok {
		fmt.Println("  \"go\" is in the set")
	}
}

// -----------------------------------------------------------------------
// SECTION 7: Maps with Slice Values (grouping)
// -----------------------------------------------------------------------
// A very common pattern: group items by some key.

func groupingWithMaps() {
	fmt.Println("\nGrouping with maps:")

	type Person struct {
		Name string
		City string
	}

	people := []Person{
		{"Alice", "NYC"},
		{"Bob", "LA"},
		{"Carol", "NYC"},
		{"Dave", "LA"},
		{"Eve", "NYC"},
	}

	// Group by city
	byCity := make(map[string][]string)
	for _, p := range people {
		byCity[p.City] = append(byCity[p.City], p.Name)
		// Note: append on a nil slice works fine here —
		// the first access returns nil, and append handles nil.
	}

	cities := []string{"NYC", "LA"}
	for _, city := range cities {
		fmt.Printf("  %s: %v\n", city, byCity[city])
	}
}

func main() {
	creatingMaps()
	nilMapGotcha()
	keyExistence()
	deletingKeys()
	iteration()
	mapsAsSets()
	groupingWithMaps()
}
