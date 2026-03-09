// Package main demonstrates advanced function features in Go.
// Topics: variadic functions, closures, function types, first-class functions.
package main

import "fmt"

// -----------------------------------------------------------------------
// SECTION 1: Multiple Return Values
// -----------------------------------------------------------------------
// Go functions can return multiple values. This is the idiomatic way to
// return a result AND an error, without exceptions.

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil
}

// Named return values — the return variables are declared in the signature.
// A bare `return` returns the current values of the named variables.
// Use sparingly — can reduce readability in long functions.
func minMax(nums []int) (min, max int) {
	if len(nums) == 0 {
		return // returns 0, 0 (zero values)
	}
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return // bare return — returns named min and max
}

func multipleReturns() {
	result, err := divide(10, 3)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Printf("10 / 3 = %.4f\n", result)
	}

	_, err = divide(5, 0) // _ discards the result we don't need
	fmt.Printf("divide by zero: %v\n", err)

	min, max := minMax([]int{3, 1, 4, 1, 5, 9, 2, 6})
	fmt.Printf("min=%d max=%d\n", min, max)
}

// -----------------------------------------------------------------------
// SECTION 2: Variadic Functions
// -----------------------------------------------------------------------
// A variadic function accepts any number of arguments for its last parameter.
// The variadic parameter is received as a SLICE inside the function.
// Declare with: func name(args ...Type)

func sum(nums ...int) int {
	// nums is []int inside the function
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Variadic with a required first parameter
func logMessage(level string, parts ...any) string {
	msg := fmt.Sprintf("[%s]", level)
	for _, p := range parts {
		msg += fmt.Sprintf(" %v", p)
	}
	return msg
}

func variadicFunctions() {
	fmt.Println("\nVariadic functions:")
	fmt.Printf("  sum(1,2,3):       %d\n", sum(1, 2, 3))
	fmt.Printf("  sum(1..10):       %d\n", sum(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
	fmt.Printf("  sum() empty:      %d\n", sum()) // variadic can receive 0 args

	// Spread a slice into a variadic call with ...
	nums := []int{10, 20, 30}
	fmt.Printf("  sum(nums...):     %d\n", sum(nums...))

	fmt.Printf("  %s\n", logMessage("INFO", "server", "started", "on port", 8080))
}

// -----------------------------------------------------------------------
// SECTION 3: Function Types
// -----------------------------------------------------------------------
// Functions are first-class values in Go. You can:
//   - Assign functions to variables
//   - Pass functions as arguments
//   - Return functions from functions

// Declare a function type for clarity
type MathFunc func(int, int) int

func apply(a, b int, fn MathFunc) int {
	return fn(a, b)
}

func functionTypes() {
	fmt.Println("\nFunction types:")

	// Assign functions to variables
	add := func(a, b int) int { return a + b }
	multiply := func(a, b int) int { return a * b }

	fmt.Printf("  apply(3, 4, add):      %d\n", apply(3, 4, add))
	fmt.Printf("  apply(3, 4, multiply): %d\n", apply(3, 4, multiply))

	// Map of function names to functions
	ops := map[string]MathFunc{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
	}

	for name, fn := range ops {
		fmt.Printf("  ops[%q](10, 3) = %d\n", name, fn(10, 3))
	}
}

// -----------------------------------------------------------------------
// SECTION 4: Closures
// -----------------------------------------------------------------------
// A closure is a function that "closes over" variables from its outer scope.
// The function captures a REFERENCE to those variables, not a copy.
// This allows the closure to read and modify them even after the outer
// function has returned.

// makeCounter returns a function that increments and returns a count.
// Each call to makeCounter creates an independent counter.
func makeCounter() func() int {
	count := 0 // this variable is captured by the returned closure
	return func() int {
		count++ // modifies the captured count
		return count
	}
}

// makeMultiplier returns a function that multiplies by n
func makeMultiplier(n int) func(int) int {
	// n is captured from the outer function's parameter
	return func(x int) int {
		return x * n
	}
}

// Practical closure: a configurable adder pipeline
func makeAdder(addend int) func(int) int {
	return func(n int) int { return n + addend }
}

func closures() {
	fmt.Println("\nClosures:")

	// Each counter is independent — separate captured `count` variable
	counter1 := makeCounter()
	counter2 := makeCounter()

	fmt.Printf("  counter1: %d %d %d\n", counter1(), counter1(), counter1())
	fmt.Printf("  counter2: %d %d\n", counter2(), counter2())
	// counter2 is unaffected by counter1 — they capture different `count` vars

	double := makeMultiplier(2)
	triple := makeMultiplier(3)
	fmt.Printf("  double(5) = %d, triple(5) = %d\n", double(5), triple(5))

	// Chain closures: add5 then double
	add5 := makeAdder(5)
	nums := []int{1, 2, 3, 4}
	for _, n := range nums {
		fmt.Printf("  add5(%d) = %d, then double = %d\n", n, add5(n), double(add5(n)))
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Closure Gotcha — Loop Variable Capture
// -----------------------------------------------------------------------
// Classic bug: capturing a loop variable by reference.
// By the time the closure runs, the loop may have finished.

func closureGotcha() {
	fmt.Println("\nClosure gotcha with loops:")

	// BUG: all closures capture the SAME `i` variable
	buggy := make([]func(), 3)
	for i := 0; i < 3; i++ {
		buggy[i] = func() {
			fmt.Printf("  buggy: i = %d\n", i) // all will print 3 (after loop ends)
		}
	}
	for _, f := range buggy {
		f()
	}

	// FIX 1: capture a copy by passing as argument
	fixed1 := make([]func(), 3)
	for i := 0; i < 3; i++ {
		i := i // new variable `i` scoped to this loop iteration
		fixed1[i] = func() {
			fmt.Printf("  fixed: i = %d\n", i)
		}
	}
	for _, f := range fixed1 {
		f()
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Immediately Invoked Function Expression (IIFE)
// -----------------------------------------------------------------------
// Create and call a function inline. Useful for scoping variables
// or running setup logic inline.

func iife() {
	fmt.Println("\nIIFE:")

	result := func(x, y int) int {
		return x * x + y * y
	}(3, 4) // immediately called with args 3 and 4

	fmt.Printf("  3^2 + 4^2 = %d\n", result)
}

func main() {
	multipleReturns()
	variadicFunctions()
	functionTypes()
	closures()
	closureGotcha()
	iife()
}
