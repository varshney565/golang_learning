// Package main demonstrates structs in Go.
// Topics: struct basics, embedding, anonymous fields, struct tags, comparison.
package main

import (
	"encoding/json"
	"fmt"
)

// -----------------------------------------------------------------------
// SECTION 1: Struct Basics
// -----------------------------------------------------------------------
// A struct groups related data together. Fields can be any type.
// Exported fields start with an uppercase letter.

type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

type Person struct {
	FirstName string
	LastName  string
	Age       int
	Address   Address // nested struct (composition)
}

func structBasics() {
	// Named field initialization — preferred, order doesn't matter
	p := Person{
		FirstName: "Alice",
		LastName:  "Smith",
		Age:       30,
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
			State:  "CA",
			Zip:    "90210",
		},
	}

	fmt.Printf("Name: %s %s, Age: %d\n", p.FirstName, p.LastName, p.Age)
	fmt.Printf("City: %s, %s\n", p.Address.City, p.Address.State)

	// Zero-value struct — all fields set to their zero values
	var empty Person
	fmt.Printf("Empty person: %+v\n", empty) // %+v shows field names
}

// -----------------------------------------------------------------------
// SECTION 2: Embedding — Go's Way of Composition
// -----------------------------------------------------------------------
// Embedding (anonymous fields) lets a struct "inherit" fields and methods
// from another type. It's NOT inheritance — it's composition.
//
// The embedded type's fields and methods are PROMOTED to the outer struct.

type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return a.Name + " makes a sound"
}

type Dog struct {
	Animal        // embedded — no field name, just the type
	Breed  string // regular field
}

// Dog can override the promoted method
func (d Dog) Speak() string {
	return d.Name + " barks" // d.Name works because Animal is embedded
}

type GuideDog struct {
	Dog            // embed Dog (which already embeds Animal)
	Handler string
}

func embedding() {
	fmt.Println("\nEmbedding:")

	dog := Dog{
		Animal: Animal{Name: "Rex"},
		Breed:  "Labrador",
	}

	// Promoted field — can access directly
	fmt.Printf("  dog.Name (promoted):  %s\n", dog.Name)
	fmt.Printf("  dog.Animal.Name:      %s\n", dog.Animal.Name) // explicit also works

	// Dog overrides Speak
	fmt.Printf("  dog.Speak():          %s\n", dog.Speak())

	// Animal.Speak is still accessible via full path
	fmt.Printf("  dog.Animal.Speak():   %s\n", dog.Animal.Speak())

	// Multi-level embedding
	guide := GuideDog{
		Dog:     dog,
		Handler: "Bob",
	}
	// Name is promoted through two levels of embedding
	fmt.Printf("  guide.Name (2-level): %s\n", guide.Name)
	fmt.Printf("  guide.Speak():        %s\n", guide.Speak()) // uses Dog's Speak
}

// -----------------------------------------------------------------------
// SECTION 3: Embedding Interfaces
// -----------------------------------------------------------------------
// You can embed interfaces in structs too. This is used heavily for
// mocking in tests — implement only the methods you care about.

type Writer interface {
	Write(data string)
}

type Logger struct {
	Writer        // embedded interface — Logger "has" a Writer
	Prefix string
}

func (l Logger) Log(msg string) {
	if l.Writer != nil {
		l.Write(l.Prefix + ": " + msg)
	}
}

// Concrete implementation
type ConsoleWriter struct{}

func (cw ConsoleWriter) Write(data string) {
	fmt.Println(" ", data)
}

func embeddingInterfaces() {
	fmt.Println("\nEmbedding interfaces:")

	log := Logger{
		Writer: ConsoleWriter{},
		Prefix: "INFO",
	}
	log.Log("server started")
}

// -----------------------------------------------------------------------
// SECTION 4: Struct Tags
// -----------------------------------------------------------------------
// Tags are metadata attached to struct fields. They're strings in backtick
// syntax and read at runtime via reflection. Common uses: JSON, DB, validation.

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`           // "-" means: NEVER marshal this field
	Email     string `json:"email,omitempty"` // omitempty: skip if zero value
	IsAdmin   bool   `json:"is_admin"`
}

func structTags() {
	fmt.Println("\nStruct tags (JSON):")

	u := User{
		ID:       1,
		Username: "alice",
		Password: "secret123",
		IsAdmin:  false,
		// Email is empty — will be omitted due to omitempty
	}

	// Marshal to JSON — tags control field names and behavior
	data, _ := json.MarshalIndent(u, "  ", "  ")
	fmt.Printf("  JSON output:\n  %s\n", data)
	// Notice: Password is absent, Email is absent (omitempty + empty)

	// Unmarshal from JSON
	jsonStr := `{"id":2,"username":"bob","email":"bob@example.com","is_admin":true}`
	var u2 User
	json.Unmarshal([]byte(jsonStr), &u2)
	fmt.Printf("  Parsed user: %+v\n", u2)
}

// -----------------------------------------------------------------------
// SECTION 5: Struct Comparison
// -----------------------------------------------------------------------
// Structs are comparable with == if all their fields are comparable.
// Slices, maps, and functions are NOT comparable → struct becomes non-comparable.

type Point struct {
	X, Y int // shorthand: multiple fields of same type on one line
}

type Segment struct {
	Start, End Point
}

func structComparison() {
	fmt.Println("\nStruct comparison:")

	p1 := Point{1, 2}
	p2 := Point{1, 2}
	p3 := Point{3, 4}

	fmt.Printf("  p1 == p2: %v (same values)\n", p1 == p2)
	fmt.Printf("  p1 == p3: %v (different values)\n", p1 == p3)

	// Nested structs — all fields must match
	s1 := Segment{Start: Point{0, 0}, End: Point{1, 1}}
	s2 := Segment{Start: Point{0, 0}, End: Point{1, 1}}
	fmt.Printf("  s1 == s2: %v\n", s1 == s2)

	// Structs can be used as map keys if they're comparable!
	cache := map[Point]string{
		{0, 0}: "origin",
		{1, 0}: "unit-x",
	}
	fmt.Printf("  cache[Point{0,0}]: %s\n", cache[Point{0, 0}])
}

// -----------------------------------------------------------------------
// SECTION 6: Anonymous Structs
// -----------------------------------------------------------------------
// Structs without a name. Useful for one-off data grouping, test cases,
// or JSON parsing into ad-hoc shapes.

func anonymousStructs() {
	fmt.Println("\nAnonymous structs:")

	// Single anonymous struct
	config := struct {
		Host string
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}
	fmt.Printf("  config: %+v\n", config)

	// Slice of anonymous structs — common in table-driven tests
	tests := []struct {
		input    int
		expected int
	}{
		{2, 4},
		{3, 9},
		{4, 16},
	}
	for _, tt := range tests {
		result := tt.input * tt.input
		status := "PASS"
		if result != tt.expected {
			status = "FAIL"
		}
		fmt.Printf("  %d^2 = %d [%s]\n", tt.input, result, status)
	}
}

func main() {
	structBasics()
	embedding()
	embeddingInterfaces()
	structTags()
	structComparison()
	anonymousStructs()
}
