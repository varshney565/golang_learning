// Package main demonstrates io and file operations in Go.
// Topics: io.Reader/Writer, bufio, os file ops, io.Copy, strings.Builder.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// -----------------------------------------------------------------------
// SECTION 1: io.Reader and io.Writer — The Core Interfaces
// -----------------------------------------------------------------------
// Almost everything that reads/writes in Go uses these interfaces:
//
//   type Reader interface { Read(p []byte) (n int, err error) }
//   type Writer interface { Write(p []byte) (n int, err error) }
//
// They compose into: ReadWriter, ReadCloser, WriteCloser, etc.
// This means the same code works for files, network, memory, HTTP bodies.

// countBytes reads from any Reader and counts bytes
func countBytes(r io.Reader) (int, error) {
	buf := make([]byte, 512) // read 512 bytes at a time
	total := 0
	for {
		n, err := r.Read(buf)
		total += n
		if err == io.EOF { // io.EOF means: no more data, not an actual error
			break
		}
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// writeTwice writes data twice to any Writer
func writeTwice(w io.Writer, data string) {
	fmt.Fprintf(w, "%s\n", data)
	fmt.Fprintf(w, "%s (again)\n", data)
}

func readerWriterDemo() {
	fmt.Println("io.Reader / io.Writer:")

	// strings.NewReader is a Reader backed by a string
	r := strings.NewReader("Hello, Go!")
	n, _ := countBytes(r)
	fmt.Printf("  counted %d bytes in \"Hello, Go!\"\n", n)

	// os.Stdout is a Writer — writeTwice works on it
	fmt.Print("  writeTwice to stdout:\n")
	writeTwice(os.Stdout, "  hello")

	// strings.Builder is a Writer — collects output in memory
	var sb strings.Builder
	writeTwice(&sb, "captured")
	fmt.Printf("  captured: %q\n", sb.String())
}

// -----------------------------------------------------------------------
// SECTION 2: io.Copy — Efficient Transfer
// -----------------------------------------------------------------------
// io.Copy(dst Writer, src Reader) copies from src to dst without loading
// everything into memory. Uses an internal buffer (32KB by default).
// Essential for streaming large files, HTTP proxying, piping.

func ioCopyDemo() {
	fmt.Println("\nio.Copy:")

	src := strings.NewReader("This is the source data being copied.")
	var dst strings.Builder

	n, err := io.Copy(&dst, src)
	fmt.Printf("  copied %d bytes, err=%v\n", n, err)
	fmt.Printf("  result: %q\n", dst.String())

	// io.TeeReader — reads from r, writes to w, returns a Reader
	// Useful for reading a body while also logging it
	var logged strings.Builder
	tee := io.TeeReader(strings.NewReader("log this"), &logged)
	data, _ := io.ReadAll(tee)
	fmt.Printf("  tee read: %q, logged: %q\n", data, logged.String())
}

// -----------------------------------------------------------------------
// SECTION 3: bufio — Buffered I/O
// -----------------------------------------------------------------------
// Raw Read/Write can be slow for small operations (syscall per call).
// bufio.Reader/Writer batches operations using an in-memory buffer.
//
// bufio.Scanner is the idiomatic way to read line by line.

func bufioDemo() {
	fmt.Println("\nbufio:")

	// bufio.Scanner — reads line by line
	input := "line 1\nline 2\nline 3\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	// scanner.Split(bufio.ScanWords)  // or scan word by word
	lineNum := 0
	for scanner.Scan() { // advances to next line, returns false at EOF
		lineNum++
		fmt.Printf("  line %d: %q\n", lineNum, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("  scanner error: %v\n", err)
	}

	// bufio.Reader — buffered reads with extra methods
	br := bufio.NewReader(strings.NewReader("key: value\nfoo: bar\n"))
	for {
		line, err := br.ReadString('\n') // read until delimiter
		if len(line) > 0 {
			fmt.Printf("  read: %q\n", strings.TrimRight(line, "\n"))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("  error: %v\n", err)
			break
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 4: File Operations
// -----------------------------------------------------------------------

func fileOperations() {
	fmt.Println("\nFile operations:")

	// Create and write a file
	tmpFile := "/tmp/go_learning_demo.txt"

	err := os.WriteFile(tmpFile, []byte("line 1\nline 2\nline 3\n"), 0644)
	if err != nil {
		fmt.Printf("  write error: %v\n", err)
		return
	}
	fmt.Printf("  wrote %s\n", tmpFile)

	// Read entire file at once (fine for small files)
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		fmt.Printf("  read error: %v\n", err)
		return
	}
	fmt.Printf("  read: %q\n", string(data))

	// Open file for streaming (better for large files)
	f, err := os.Open(tmpFile) // read-only
	if err != nil {
		fmt.Printf("  open error: %v\n", err)
		return
	}
	defer f.Close() // always close

	// Get file info
	info, _ := f.Stat()
	fmt.Printf("  size: %d bytes, mode: %v\n", info.Size(), info.Mode())

	// Scan lines
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Printf("  > %s\n", scanner.Text())
	}

	// Cleanup
	os.Remove(tmpFile)
}

// -----------------------------------------------------------------------
// SECTION 5: Create, Append, Write with os.OpenFile
// -----------------------------------------------------------------------

func openFileFlags() {
	fmt.Println("\nos.OpenFile flags:")
	fmt.Println("  os.O_RDONLY — read only")
	fmt.Println("  os.O_WRONLY — write only")
	fmt.Println("  os.O_RDWR   — read + write")
	fmt.Println("  os.O_CREATE — create if not exists")
	fmt.Println("  os.O_TRUNC  — truncate on open")
	fmt.Println("  os.O_APPEND — append to end")
	fmt.Println()

	tmpFile := "/tmp/go_append_demo.txt"
	defer os.Remove(tmpFile)

	// Append to file (create if not exists)
	f, err := os.OpenFile(tmpFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}

	bw := bufio.NewWriter(f)
	fmt.Fprintln(bw, "appended line 1")
	fmt.Fprintln(bw, "appended line 2")
	bw.Flush() // IMPORTANT: flush buffered writer before close
	f.Close()

	data, _ := os.ReadFile(tmpFile)
	fmt.Printf("  file contents:\n%s", data)
}

// -----------------------------------------------------------------------
// SECTION 6: strings and bytes — Common io.Reader/Writer Helpers
// -----------------------------------------------------------------------

func ioHelpers() {
	fmt.Println("\nio helpers:")

	// strings.NewReader — string as Reader
	r := strings.NewReader("hello world")
	fmt.Printf("  strings.NewReader len=%d\n", r.Len())

	// io.ReadAll — read everything from a Reader into []byte
	all, _ := io.ReadAll(strings.NewReader("read all of this"))
	fmt.Printf("  io.ReadAll: %q\n", all)

	// io.LimitReader — wrap a reader with a byte limit
	limited := io.LimitReader(strings.NewReader("long string here"), 4)
	data, _ := io.ReadAll(limited)
	fmt.Printf("  LimitReader(4): %q\n", data)

	// io.MultiReader — chain multiple readers
	r1 := strings.NewReader("part1-")
	r2 := strings.NewReader("part2-")
	r3 := strings.NewReader("part3")
	combined := io.MultiReader(r1, r2, r3)
	data, _ = io.ReadAll(combined)
	fmt.Printf("  MultiReader: %q\n", data)

	// io.MultiWriter — fan-out to multiple writers
	var a, b strings.Builder
	mw := io.MultiWriter(&a, &b)
	fmt.Fprint(mw, "written to both")
	fmt.Printf("  MultiWriter a=%q b=%q\n", a.String(), b.String())
}

func main() {
	readerWriterDemo()
	ioCopyDemo()
	bufioDemo()
	fileOperations()
	openFileFlags()
	ioHelpers()
}
