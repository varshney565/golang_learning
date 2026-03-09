// Package main demonstrates slices in Go — one of the most important data structures.
// Topics: slice internals (len/cap), append behavior, copy, 2D slices, gotchas.
package main

import "fmt"

// -----------------------------------------------------------------------
// SECTION 1: Slice Internals
// -----------------------------------------------------------------------
// A slice is NOT an array. It's a lightweight struct containing:
//   - Pointer → to an underlying array
//   - Length  → number of elements currently visible
//   - Capacity → total space in the underlying array from the pointer
//
// Multiple slices can share the same underlying array!

func sliceInternals() {
	// Array — fixed size, value type
	arr := [5]int{10, 20, 30, 40, 50}

	// Slice from array — shares the same memory
	s := arr[1:4] // elements at index 1, 2, 3 → [20, 30, 40]
	fmt.Printf("arr:    %v\n", arr)
	fmt.Printf("s:      %v  len=%d  cap=%d\n", s, len(s), cap(s))
	// cap=4 because from index 1 to end of arr there are 4 slots

	// Modifying slice modifies the underlying array
	s[0] = 999
	fmt.Printf("after s[0]=999: arr=%v\n", arr) // arr[1] is now 999
}

// -----------------------------------------------------------------------
// SECTION 2: Making Slices
// -----------------------------------------------------------------------

func makingSlices() {
	fmt.Println("\nMaking slices:")

	// Literal — length and capacity both = number of elements
	s1 := []int{1, 2, 3}
	fmt.Printf("  literal:     %v len=%d cap=%d\n", s1, len(s1), cap(s1))

	// make([]T, length, capacity) — allocates backing array
	// Use make when you know the size upfront — avoids repeated reallocations
	s2 := make([]int, 3, 5) // 3 elements (all 0), capacity for 5
	fmt.Printf("  make(3,5):   %v len=%d cap=%d\n", s2, len(s2), cap(s2))

	// Nil slice — zero value, no backing array, len=0 cap=0
	// Behaves just like an empty slice for most purposes
	var s3 []int
	fmt.Printf("  nil slice:   %v len=%d cap=%d nil=%v\n", s3, len(s3), cap(s3), s3 == nil)
}

// -----------------------------------------------------------------------
// SECTION 3: append — The Critical Part
// -----------------------------------------------------------------------
// append adds elements to a slice. The tricky part:
//   - If cap is sufficient → appends in place, same backing array
//   - If cap is exceeded   → allocates NEW backing array (typically 2x), copies data
//
// This means: ALWAYS reassign the result of append back to your variable.

func appendBehavior() {
	fmt.Println("\nappend behavior:")

	s := make([]int, 0, 3) // empty, but capacity for 3
	fmt.Printf("  start:       %v len=%d cap=%d\n", s, len(s), cap(s))

	// These appends fit within cap — no reallocation
	s = append(s, 1)
	fmt.Printf("  append 1:    %v len=%d cap=%d\n", s, len(s), cap(s))
	s = append(s, 2)
	s = append(s, 3)
	fmt.Printf("  append 2,3:  %v len=%d cap=%d\n", s, len(s), cap(s))

	// This exceeds cap — Go allocates a new, larger backing array
	s = append(s, 4)
	fmt.Printf("  append 4:    %v len=%d cap=%d (NEW backing array)\n", s, len(s), cap(s))

	// Append multiple elements at once with ...
	extra := []int{5, 6, 7}
	s = append(s, extra...) // spread the slice
	fmt.Printf("  append ...%v: %v\n", extra, s)
}

// -----------------------------------------------------------------------
// SECTION 4: The Shared Backing Array Gotcha
// -----------------------------------------------------------------------
// Because slices share backing arrays, mutations can have surprising effects.

func sharedArrayGotcha() {
	fmt.Println("\nShared backing array gotcha:")

	original := []int{1, 2, 3, 4, 5}
	slice1 := original[0:3] // [1, 2, 3] — shares original's array
	slice2 := original[0:3] // [1, 2, 3] — also shares original's array

	// Modifying slice1 also changes slice2 and original!
	slice1[0] = 999
	fmt.Printf("  original: %v\n", original) // [999 2 3 4 5]
	fmt.Printf("  slice1:   %v\n", slice1)   // [999 2 3 4 5]
	fmt.Printf("  slice2:   %v\n", slice2)   // [999 2 3 4 5] ← also changed!
}

// -----------------------------------------------------------------------
// SECTION 5: copy — Break the Shared Reference
// -----------------------------------------------------------------------
// copy(dst, src) copies min(len(dst), len(src)) elements.
// The two slices no longer share a backing array.

func copySlice() {
	fmt.Println("\ncopy to break shared reference:")

	original := []int{1, 2, 3, 4, 5}

	// Make an independent copy
	clone := make([]int, len(original))
	n := copy(clone, original)
	fmt.Printf("  copied %d elements\n", n)

	clone[0] = 999
	fmt.Printf("  original: %v (unchanged)\n", original)
	fmt.Printf("  clone:    %v\n", clone)
}

// -----------------------------------------------------------------------
// SECTION 6: Deleting from a Slice
// -----------------------------------------------------------------------
// There's no built-in delete for slices. Common patterns:

func deletingElements() {
	fmt.Println("\nDeleting elements:")

	s := []int{1, 2, 3, 4, 5}

	// Delete element at index 2 (value 3) — order preserved
	i := 2
	s = append(s[:i], s[i+1:]...)
	fmt.Printf("  after deleting index 2: %v\n", s)

	// Delete element at index 1 — order NOT preserved (faster)
	s2 := []int{1, 2, 3, 4, 5}
	j := 1
	s2[j] = s2[len(s2)-1] // replace with last
	s2 = s2[:len(s2)-1]   // shrink by 1
	fmt.Printf("  fast delete index 1:    %v\n", s2)
}

// -----------------------------------------------------------------------
// SECTION 7: 2D Slices
// -----------------------------------------------------------------------

func twoDimensional() {
	fmt.Println("\n2D slices (slice of slices):")

	// Create a 3x4 grid
	rows, cols := 3, 4
	grid := make([][]int, rows)
	for i := range grid {
		grid[i] = make([]int, cols)
		for j := range grid[i] {
			grid[i][j] = i*cols + j
		}
	}

	for _, row := range grid {
		fmt.Printf("  %v\n", row)
	}
}

// -----------------------------------------------------------------------
// SECTION 8: Slice Tricks Summary
// -----------------------------------------------------------------------

func sliceTricks() {
	fmt.Println("\nSlice tricks:")

	s := []int{1, 2, 3, 4, 5}

	// Reverse in place
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	fmt.Printf("  reversed: %v\n", s)

	// Filter (keep evens) — no allocation trick
	result := s[:0] // same backing array, len=0
	for _, v := range s {
		if v%2 == 0 {
			result = append(result, v)
		}
	}
	fmt.Printf("  evens:    %v\n", result)
}

func main() {
	sliceInternals()
	makingSlices()
	appendBehavior()
	sharedArrayGotcha()
	copySlice()
	deletingElements()
	twoDimensional()
	sliceTricks()
}
