// Package main demonstrates JSON encoding/decoding in Go.
// Topics: marshal/unmarshal, tags, custom marshalers, streaming, unknown fields.
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Basic Marshal / Unmarshal
// -----------------------------------------------------------------------
// json.Marshal   → Go value → JSON bytes
// json.Unmarshal → JSON bytes → Go value
//
// json.MarshalIndent → pretty-printed JSON

type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
	Zip    string `json:"zip"`
}

type Person struct {
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Email   string  `json:"email,omitempty"` // omit if empty string
	Score   float64 `json:"score,omitempty"` // omit if 0
	Address Address `json:"address"`
	Tags    []string `json:"tags,omitempty"` // omit if nil/empty slice
}

func basicMarshal() {
	fmt.Println("Basic marshal/unmarshal:")

	p := Person{
		Name: "Alice",
		Age:  30,
		// Email is empty — will be omitted
		Address: Address{Street: "123 Main", City: "Springfield", Zip: "12345"},
		Tags:    []string{"admin", "user"},
	}

	// Compact JSON
	data, err := json.Marshal(p)
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	fmt.Printf("  compact: %s\n", data)

	// Pretty-printed JSON
	pretty, _ := json.MarshalIndent(p, "  ", "  ")
	fmt.Printf("  pretty:\n  %s\n", pretty)

	// Unmarshal back
	jsonStr := `{"name":"Bob","age":25,"email":"bob@example.com","address":{"street":"456 Oak","city":"Shelbyville","zip":"67890"}}`
	var p2 Person
	if err := json.Unmarshal([]byte(jsonStr), &p2); err != nil {
		fmt.Printf("  unmarshal error: %v\n", err)
		return
	}
	fmt.Printf("  unmarshalled: %+v\n", p2)
}

// -----------------------------------------------------------------------
// SECTION 2: JSON Tags Deep Dive
// -----------------------------------------------------------------------
// Tag format: `json:"fieldname,options"`
// Options:
//   omitempty — skip field if zero value (empty string, 0, false, nil, empty slice/map)
//   -         — always skip this field
//   string    — encode number/bool as a JSON string

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port,string"`  // encodes as "8080" (string) not 8080 (number)
	Debug    bool   `json:"debug"`
	Password string `json:"-"`            // never include in JSON
	Internal string `json:"-,"`           // field named "-" (rare)
}

func jsonTags() {
	fmt.Println("\nJSON tags:")

	c := Config{Host: "localhost", Port: 8080, Debug: true, Password: "secret"}
	data, _ := json.MarshalIndent(c, "  ", "  ")
	fmt.Printf("  %s\n", data)
	// Notice: Password is absent, Port is "8080" (string)

	// Unmarshal with string-encoded number
	jsonStr := `{"host":"prod.example.com","port":"9090","debug":false}`
	var c2 Config
	json.Unmarshal([]byte(jsonStr), &c2)
	fmt.Printf("  parsed port: %d (int from string)\n", c2.Port)
}

// -----------------------------------------------------------------------
// SECTION 3: Custom Marshaler / Unmarshaler
// -----------------------------------------------------------------------
// Implement json.Marshaler / json.Unmarshaler interfaces for custom behavior.

type Duration struct {
	time.Duration // embed — gets Duration methods
}

// MarshalJSON encodes Duration as a human-readable string: "1h30m"
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String()) // e.g., "1h30m0s"
}

// UnmarshalJSON parses a human-readable string back to Duration
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dur
	return nil
}

type ServerConfig struct {
	Timeout Duration `json:"timeout"`
	Retry   int      `json:"retry"`
}

func customMarshaler() {
	fmt.Println("\nCustom marshaler:")

	cfg := ServerConfig{
		Timeout: Duration{90 * time.Second},
		Retry:   3,
	}
	data, _ := json.Marshal(cfg)
	fmt.Printf("  marshalled: %s\n", data)

	var cfg2 ServerConfig
	json.Unmarshal([]byte(`{"timeout":"2m30s","retry":5}`), &cfg2)
	fmt.Printf("  parsed timeout: %v (%d seconds)\n", cfg2.Timeout, int(cfg2.Timeout.Seconds()))
}

// -----------------------------------------------------------------------
// SECTION 4: Handling Unknown / Dynamic Fields
// -----------------------------------------------------------------------

// Map approach — for fully dynamic JSON
func dynamicJSON() {
	fmt.Println("\nDynamic JSON (map):")

	jsonStr := `{"name":"Alice","role":"admin","extra_field":42,"nested":{"key":"val"}}`

	var m map[string]any
	json.Unmarshal([]byte(jsonStr), &m)

	for k, v := range m {
		fmt.Printf("  %s: %v (%T)\n", k, v, v)
	}
}

// RawMessage — defer parsing of a field
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"` // raw bytes, parse later
}

type ClickPayload struct {
	X, Y int `json:"x,y"`
}

type KeyPayload struct {
	Key string `json:"key"`
}

func rawMessageDemo() {
	fmt.Println("\njson.RawMessage (defer parsing):")

	events := []string{
		`{"type":"click","payload":{"x":100,"y":200}}`,
		`{"type":"keypress","payload":{"key":"Enter"}}`,
	}

	for _, raw := range events {
		var event Event
		json.Unmarshal([]byte(raw), &event)

		switch event.Type {
		case "click":
			var p ClickPayload
			json.Unmarshal(event.Payload, &p)
			fmt.Printf("  click at (%d, %d)\n", p.X, p.Y)
		case "keypress":
			var p KeyPayload
			json.Unmarshal(event.Payload, &p)
			fmt.Printf("  key pressed: %q\n", p.Key)
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Streaming JSON — Encoder / Decoder
// -----------------------------------------------------------------------
// json.NewEncoder/Decoder work with io.Reader/Writer.
// Use for:
//   - Large JSON files (no need to load entirely into memory)
//   - HTTP request/response bodies
//   - NDJSON (newline-delimited JSON)

func streamingJSON() {
	fmt.Println("\nStreaming JSON (Encoder/Decoder):")

	people := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Carol", Age: 35},
	}

	// Encode directly to a strings.Builder (simulating a network write)
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ") // optional pretty-print

	for _, p := range people {
		enc.Encode(p) // writes one JSON object per line
	}
	fmt.Printf("  encoded:\n%s\n", buf.String())

	// Decode from a Reader (simulating reading from network/file)
	dec := json.NewDecoder(strings.NewReader(buf.String()))
	for dec.More() { // More() returns true if there's another value
		var p Person
		if err := dec.Decode(&p); err != nil {
			fmt.Printf("  decode error: %v\n", err)
			break
		}
		fmt.Printf("  decoded: %s (age %d)\n", p.Name, p.Age)
	}
}

// -----------------------------------------------------------------------
// SECTION 6: DisallowUnknownFields
// -----------------------------------------------------------------------
// By default, JSON decoder ignores unknown fields.
// Use DisallowUnknownFields() to return an error on unknown fields.
// Useful for strict API validation.

func disallowUnknown() {
	fmt.Println("\nDisallowUnknownFields:")

	type Strict struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	goodJSON := `{"name":"Alice","age":30}`
	badJSON := `{"name":"Bob","age":25,"unknown_field":"oops"}`

	for _, jsonStr := range []string{goodJSON, badJSON} {
		var s Strict
		dec := json.NewDecoder(strings.NewReader(jsonStr))
		dec.DisallowUnknownFields()

		if err := dec.Decode(&s); err != nil {
			fmt.Printf("  error: %v\n", err)
		} else {
			fmt.Printf("  ok: %+v\n", s)
		}
	}
}

func main() {
	basicMarshal()
	jsonTags()
	customMarshaler()
	dynamicJSON()
	rawMessageDemo()
	streamingJSON()
	disallowUnknown()
}
