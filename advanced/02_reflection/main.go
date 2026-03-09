// Package main demonstrates reflection in Go.
// Topics: reflect.Type, reflect.Value, struct inspection, dynamic calls.
// Reflection is an interview topic — know when NOT to use it too.
package main

import (
	"fmt"
	"reflect"
)

// -----------------------------------------------------------------------
// SECTION 1: reflect.Type and reflect.Value
// -----------------------------------------------------------------------
// reflect package gives you runtime access to type information.
//
//   reflect.TypeOf(x)  → reflect.Type  (metadata about the type)
//   reflect.ValueOf(x) → reflect.Value (the value itself, at runtime)

func typeAndValue() {
	fmt.Println("reflect.Type and reflect.Value:")

	values := []any{42, 3.14, "hello", true, []int{1, 2, 3}}

	for _, v := range values {
		t := reflect.TypeOf(v)
		rv := reflect.ValueOf(v)
		fmt.Printf("  %-20T kind=%-10s type=%s\n", v, rv.Kind(), t.Name())
		// Kind is the category (int, slice, struct...) — not the specific type name
	}
}

// -----------------------------------------------------------------------
// SECTION 2: Kind vs Type
// -----------------------------------------------------------------------
// Type: the specific named type (e.g., "MyInt", "Person")
// Kind: the category (e.g., int, struct, slice, ptr)
//
// Custom types have the same Kind as their underlying type.

type MyInt int
type Person struct{ Name string }

func kindVsType() {
	fmt.Println("\nKind vs Type:")

	var mi MyInt = 5
	var p Person

	fmt.Printf("  MyInt:  Type=%s Kind=%s\n", reflect.TypeOf(mi).Name(), reflect.TypeOf(mi).Kind())
	fmt.Printf("  Person: Type=%s Kind=%s\n", reflect.TypeOf(p).Name(), reflect.TypeOf(p).Kind())
	fmt.Printf("  *Person:Type=%s Kind=%s\n", reflect.TypeOf(&p).Name(), reflect.TypeOf(&p).Kind())
}

// -----------------------------------------------------------------------
// SECTION 3: Inspecting Structs
// -----------------------------------------------------------------------

type Employee struct {
	Name       string `json:"name" validate:"required"`
	Age        int    `json:"age"  validate:"min=18"`
	Department string `json:"dept" validate:"required"`
	salary     float64 // unexported
}

func inspectStruct(v any) {
	t := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	// Dereference pointer if needed
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		rv = rv.Elem()
	}

	fmt.Printf("  Struct: %s\n", t.Name())
	fmt.Printf("  Fields (%d):\n", t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := rv.Field(i)

		// CanInterface() returns false for unexported fields
		if !value.CanInterface() {
			fmt.Printf("    %-15s [unexported]\n", field.Name)
			continue
		}

		fmt.Printf("    %-15s type=%-10s value=%-15v json=%q validate=%q\n",
			field.Name,
			field.Type,
			value.Interface(),
			field.Tag.Get("json"),
			field.Tag.Get("validate"),
		)
	}
}

func structInspection() {
	fmt.Println("\nStruct inspection:")
	e := Employee{Name: "Alice", Age: 30, Department: "Engineering"}
	inspectStruct(e)
}

// -----------------------------------------------------------------------
// SECTION 4: Modifying Values via Reflection
// -----------------------------------------------------------------------
// To modify a value via reflection, you must pass a POINTER.
// Use CanSet() to check if a field is settable.

func modifyViaReflection() {
	fmt.Println("\nModifying via reflection:")

	e := Employee{Name: "Bob", Age: 25}
	rv := reflect.ValueOf(&e).Elem() // Elem() to get the struct, not the pointer

	nameField := rv.FieldByName("Name")
	if nameField.CanSet() {
		nameField.SetString("Carol")
	}

	ageField := rv.FieldByName("Age")
	if ageField.CanSet() {
		ageField.SetInt(35)
	}

	// Can't set unexported field
	salaryField := rv.FieldByName("salary")
	fmt.Printf("  salary.CanSet(): %v\n", salaryField.CanSet())

	fmt.Printf("  modified: %+v\n", e)
}

// -----------------------------------------------------------------------
// SECTION 5: Calling Methods Dynamically
// -----------------------------------------------------------------------

type Calculator struct{}

func (c Calculator) Add(a, b int) int      { return a + b }
func (c Calculator) Multiply(a, b int) int { return a * b }

func callMethodDynamically() {
	fmt.Println("\nDynamic method calls:")

	calc := Calculator{}
	rv := reflect.ValueOf(calc)

	methods := []struct {
		name string
		args []int
	}{
		{"Add", []int{3, 4}},
		{"Multiply", []int{5, 6}},
	}

	for _, m := range methods {
		method := rv.MethodByName(m.name)
		if !method.IsValid() {
			fmt.Printf("  method %s not found\n", m.name)
			continue
		}

		// Convert args to []reflect.Value
		args := make([]reflect.Value, len(m.args))
		for i, arg := range m.args {
			args[i] = reflect.ValueOf(arg)
		}

		// Call and get results
		results := method.Call(args)
		fmt.Printf("  %s(%v) = %v\n", m.name, m.args, results[0].Interface())
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Practical Use — Deep Copy
// -----------------------------------------------------------------------
// Reflection allows writing generic utilities like deep copy, comparison, etc.

func deepCopy(src any) any {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		// Create new pointer, copy pointed-to value
		newPtr := reflect.New(srcVal.Elem().Type())
		newPtr.Elem().Set(srcVal.Elem())
		return newPtr.Interface()
	}
	// For non-pointers, just copy the value
	newVal := reflect.New(srcVal.Type()).Elem()
	newVal.Set(srcVal)
	return newVal.Interface()
}

func deepCopyDemo() {
	fmt.Println("\nDeep copy via reflection:")

	original := &Employee{Name: "Dave", Age: 40, Department: "HR"}
	copied := deepCopy(original).(*Employee)

	// Modify copy — original should be unchanged
	copied.Name = "Modified"

	fmt.Printf("  original: %+v\n", original)
	fmt.Printf("  copied:   %+v\n", copied)
}

// -----------------------------------------------------------------------
// SECTION 7: Practical Use — Simple Validator
// -----------------------------------------------------------------------
// A minimal validator that reads "required" tag and checks zero values.

func validateRequired(v any) []string {
	var errors []string
	t := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		rv = rv.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := rv.Field(i)

		tag := field.Tag.Get("validate")
		if tag == "required" || contains(tag, "required") {
			// Check if zero value
			zero := reflect.Zero(field.Type)
			if reflect.DeepEqual(value.Interface(), zero.Interface()) {
				errors = append(errors, fmt.Sprintf("field %q is required", field.Name))
			}
		}
	}
	return errors
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 &&
		(s[:len(sub)] == sub || len(s) > len(sub) && contains(s[1:], sub)))
}

func validatorDemo() {
	fmt.Println("\nSimple validator (using reflection):")

	valid := Employee{Name: "Alice", Age: 30, Department: "Eng"}
	invalid := Employee{Name: "", Age: 30} // Department is missing

	if errs := validateRequired(valid); len(errs) > 0 {
		fmt.Printf("  valid struct errors: %v\n", errs)
	} else {
		fmt.Println("  valid struct: OK")
	}

	if errs := validateRequired(invalid); len(errs) > 0 {
		for _, e := range errs {
			fmt.Printf("  invalid: %s\n", e)
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 8: When NOT to Use Reflection
// -----------------------------------------------------------------------
// Reflection is:
//   - Slow (10-100x slower than direct code)
//   - Bypasses type safety at compile time
//   - Hard to read and maintain
//
// Prefer reflection for:
//   - Serialization libraries (json, yaml, orm)
//   - Dependency injection frameworks
//   - Generic testing utilities
//   - Code generation bootstrapping
//
// Avoid reflection when generics or interfaces solve the problem.

func main() {
	typeAndValue()
	kindVsType()
	structInspection()
	modifyViaReflection()
	callMethodDynamically()
	deepCopyDemo()
	validatorDemo()
}
