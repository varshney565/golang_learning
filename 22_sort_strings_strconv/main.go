// Package main demonstrates sort, strings, and strconv packages.
// These are bread-and-butter packages asked in almost every Go interview.
package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// =======================================================================
// PART 1: sort package
// =======================================================================

// -----------------------------------------------------------------------
// SECTION 1: sort.Slice — Sort Anything
// -----------------------------------------------------------------------
// sort.Slice(slice, less func(i, j int) bool)
// You provide the "less than" function; sort handles the rest.
// NOT stable — use sort.SliceStable if equal elements must keep their order.

type Person struct {
	Name string
	Age  int
}

func sortSlice() {
	fmt.Println("sort.Slice:")

	people := []Person{
		{"Charlie", 30},
		{"Alice", 25},
		{"Bob", 35},
		{"Dave", 25},
	}

	// Sort by Age ascending
	sort.Slice(people, func(i, j int) bool {
		return people[i].Age < people[j].Age
	})
	fmt.Printf("  by age asc:  %v\n", people)

	// Sort by Age desc, then Name asc for ties
	sort.SliceStable(people, func(i, j int) bool {
		if people[i].Age != people[j].Age {
			return people[i].Age > people[j].Age
		}
		return people[i].Name < people[j].Name
	})
	fmt.Printf("  by age desc, name asc: %v\n", people)
}

// -----------------------------------------------------------------------
// SECTION 2: sort Built-in Helpers
// -----------------------------------------------------------------------

func sortBuiltins() {
	fmt.Println("\nsort built-ins:")

	// sort.Ints, sort.Strings, sort.Float64s
	nums := []int{5, 2, 8, 1, 9, 3}
	sort.Ints(nums)
	fmt.Printf("  sort.Ints:    %v\n", nums)

	words := []string{"banana", "apple", "cherry", "date"}
	sort.Strings(words)
	fmt.Printf("  sort.Strings: %v\n", words)

	// sort.Reverse — wraps any sort.Interface to reverse order
	sort.Sort(sort.Reverse(sort.IntSlice(nums)))
	fmt.Printf("  reversed:     %v\n", nums)

	// sort.Search — binary search (slice must be sorted)
	// Returns smallest index i in [0,n) where f(i) is true
	nums2 := []int{1, 3, 5, 7, 9, 11}
	target := 7
	idx := sort.SearchInts(nums2, target)
	fmt.Printf("  SearchInts(%v, %d) = index %d\n", nums2, target, idx)

	// Generic binary search
	idx2 := sort.Search(len(nums2), func(i int) bool {
		return nums2[i] >= target
	})
	fmt.Printf("  sort.Search  = index %d, value %d\n", idx2, nums2[idx2])
}

// -----------------------------------------------------------------------
// SECTION 3: Implementing sort.Interface
// -----------------------------------------------------------------------
// For custom types, implement sort.Interface:
//   Len() int, Less(i, j int) bool, Swap(i, j int)

type ByLength []string

func (b ByLength) Len() int           { return len(b) }
func (b ByLength) Less(i, j int) bool { return len(b[i]) < len(b[j]) }
func (b ByLength) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func sortInterface() {
	fmt.Println("\nsort.Interface:")
	words := []string{"banana", "kiwi", "apple", "fig", "blueberry"}
	sort.Sort(ByLength(words))
	fmt.Printf("  by length: %v\n", words)
}

// =======================================================================
// PART 2: strings package
// =======================================================================

func stringsPackage() {
	fmt.Println("\n--- strings package ---")

	s := "  Hello, Go World!  "

	// ── Basic checks ──────────────────────────────────────────
	fmt.Printf("Contains \"Go\":      %v\n", strings.Contains(s, "Go"))
	fmt.Printf("HasPrefix \"  Hello\": %v\n", strings.HasPrefix(s, "  Hello"))
	fmt.Printf("HasSuffix \"!  \":    %v\n", strings.HasSuffix(s, "!  "))
	fmt.Printf("Count \"o\":          %v\n", strings.Count(s, "o"))
	fmt.Printf("Index \"Go\":         %v\n", strings.Index(s, "Go"))
	fmt.Printf("ContainsAny \"aeiou\":%v\n", strings.ContainsAny(s, "aeiou"))

	// ── Transformation ───────────────────────────────────────
	fmt.Printf("\nToUpper:   %q\n", strings.ToUpper("hello"))
	fmt.Printf("ToLower:   %q\n", strings.ToLower("HELLO"))
	fmt.Printf("Title:     %q\n", strings.Title("hello world")) // deprecated, use cases pkg
	fmt.Printf("TrimSpace: %q\n", strings.TrimSpace(s))
	fmt.Printf("Trim:      %q\n", strings.Trim("--hello--", "-"))
	fmt.Printf("TrimLeft:  %q\n", strings.TrimLeft("--hello--", "-"))
	fmt.Printf("TrimPrefix:%q\n", strings.TrimPrefix("Go1.21", "Go"))
	fmt.Printf("TrimSuffix:%q\n", strings.TrimSuffix("main.go", ".go"))

	// ── Split / Join ──────────────────────────────────────────
	fmt.Printf("\nSplit:     %v\n", strings.Split("a,b,c", ","))
	fmt.Printf("Fields:    %v\n", strings.Fields("  foo bar  baz  ")) // splits on whitespace
	fmt.Printf("Join:      %q\n", strings.Join([]string{"a", "b", "c"}, "-"))
	// SplitN — split into at most N parts
	fmt.Printf("SplitN:    %v\n", strings.SplitN("a:b:c:d", ":", 2))

	// ── Replace ───────────────────────────────────────────────
	fmt.Printf("\nReplace:   %q\n", strings.Replace("aabbcc", "b", "X", 1))  // replace first 1
	fmt.Printf("ReplaceAll:%q\n", strings.ReplaceAll("aabbcc", "b", "X")) // replace all

	// ── Repeat ───────────────────────────────────────────────
	fmt.Printf("Repeat:    %q\n", strings.Repeat("ab", 3))

	// ── strings.Builder — efficient concatenation ─────────────
	var sb strings.Builder
	words := []string{"one", "two", "three"}
	for i, w := range words {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(w)
	}
	fmt.Printf("\nBuilder:   %q\n", sb.String())

	// ── strings.Reader — string as io.Reader ──────────────────
	r := strings.NewReader("hello")
	fmt.Printf("Reader len: %d\n", r.Len())

	// ── strings.Map — transform each rune ────────────────────
	rot13 := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		}
		return r
	}
	fmt.Printf("ROT13:     %q\n", strings.Map(rot13, "Hello, World!"))

	// ── EqualFold — case-insensitive compare ─────────────────
	fmt.Printf("EqualFold: %v\n", strings.EqualFold("Go", "go"))

	// ── ContainsRune, IndexRune ───────────────────────────────
	fmt.Printf("ContainsRune '🚀': %v\n", strings.ContainsRune("Go 🚀", '🚀'))

	// ── Cut (Go 1.18+) — split on FIRST occurrence ───────────
	before, after, found := strings.Cut("user@example.com", "@")
	fmt.Printf("Cut(@):    before=%q after=%q found=%v\n", before, after, found)
}

// =======================================================================
// PART 3: strconv package
// =======================================================================

func strconvPackage() {
	fmt.Println("\n--- strconv package ---")

	// ── Atoi / Itoa — int ↔ string (most common) ─────────────
	n, err := strconv.Atoi("42")
	fmt.Printf("Atoi(\"42\"):       %d err=%v\n", n, err)

	_, err = strconv.Atoi("abc")
	fmt.Printf("Atoi(\"abc\"):      err=%v\n", err)
	// err is a *strconv.NumError with fields: Func, Num, Err
	if numErr, ok := err.(*strconv.NumError); ok {
		fmt.Printf("  NumError: Func=%s Num=%s Err=%v\n", numErr.Func, numErr.Num, numErr.Err)
	}

	fmt.Printf("Itoa(123):        %q\n", strconv.Itoa(123))

	// ── ParseInt / FormatInt — base control ──────────────────
	// ParseInt(s, base, bitSize)
	hex, _ := strconv.ParseInt("FF", 16, 64) // base 16
	bin, _ := strconv.ParseInt("1010", 2, 64) // base 2
	fmt.Printf("ParseInt(FF,16):  %d\n", hex)
	fmt.Printf("ParseInt(1010,2): %d\n", bin)

	fmt.Printf("FormatInt(255,16):%s\n", strconv.FormatInt(255, 16)) // "ff"
	fmt.Printf("FormatInt(10,2):  %s\n", strconv.FormatInt(10, 2))   // "1010"

	// ── ParseFloat / FormatFloat ──────────────────────────────
	f, _ := strconv.ParseFloat("3.14159", 64)
	fmt.Printf("ParseFloat:       %f\n", f)
	fmt.Printf("FormatFloat:      %s\n", strconv.FormatFloat(f, 'f', 2, 64)) // "3.14"
	fmt.Printf("FormatFloat (e):  %s\n", strconv.FormatFloat(f, 'e', 3, 64)) // scientific

	// ── ParseBool / FormatBool ────────────────────────────────
	b1, _ := strconv.ParseBool("true")
	b2, _ := strconv.ParseBool("1")
	b3, _ := strconv.ParseBool("T")
	fmt.Printf("ParseBool:        %v %v %v\n", b1, b2, b3)
	fmt.Printf("FormatBool:       %s\n", strconv.FormatBool(true))

	// ── AppendInt / AppendFloat — zero-allocation versions ───
	// Append directly to a byte slice instead of creating a new string
	buf := make([]byte, 0, 20)
	buf = strconv.AppendInt(buf, 42, 10)
	buf = append(buf, ' ')
	buf = strconv.AppendFloat(buf, 3.14, 'f', 2, 64)
	fmt.Printf("Append methods:   %s\n", buf)

	// ── Quote / Unquote ───────────────────────────────────────
	fmt.Printf("Quote:    %s\n", strconv.Quote(`Hello "World"`))
	unq, _ := strconv.Unquote(`"Hello \"World\""`)
	fmt.Printf("Unquote:  %s\n", unq)
}

// =======================================================================
// PART 4: unicode package (bonus)
// =======================================================================

func unicodePackage() {
	fmt.Println("\n--- unicode package ---")

	runes := []rune{'A', 'a', '5', ' ', '!', 'é', '中'}
	for _, r := range runes {
		fmt.Printf("  %c: IsLetter=%v IsDigit=%v IsUpper=%v IsSpace=%v\n",
			r,
			unicode.IsLetter(r),
			unicode.IsDigit(r),
			unicode.IsUpper(r),
			unicode.IsSpace(r),
		)
	}

	fmt.Printf("  ToUpper('a'): %c\n", unicode.ToUpper('a'))
	fmt.Printf("  ToLower('A'): %c\n", unicode.ToLower('A'))
}

func main() {
	// sort
	sortSlice()
	sortBuiltins()
	sortInterface()

	// strings
	stringsPackage()

	// strconv
	strconvPackage()

	// unicode
	unicodePackage()
}
