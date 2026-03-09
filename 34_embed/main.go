// Package main demonstrates //go:embed — embedding files into the binary.
// go:embed was added in Go 1.16 and is asked about in interviews.
// Topics: embed single file, directory, fs.FS, use cases.
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

// -----------------------------------------------------------------------
// SECTION 1: What is //go:embed?
// -----------------------------------------------------------------------
// //go:embed lets you include files/directories INTO the compiled binary.
// No more shipping config files or static assets separately.
//
// Supported variable types:
//   string       — file contents as a string
//   []byte       — file contents as bytes
//   embed.FS     — a virtual filesystem (for multiple files/directories)
//
// The directive must appear IMMEDIATELY before the variable declaration.
// The path is relative to the source file's directory.

// -----------------------------------------------------------------------
// SECTION 2: Embedding a Single File
// -----------------------------------------------------------------------

// Create a fake "file" using a string builder since we can't embed in demos
// In real code this would be:
//   //go:embed version.txt
//   var version string

// For demonstration, we'll simulate what embedded content looks like
var version = "1.2.3" // in real code: //go:embed version.txt

// -----------------------------------------------------------------------
// SECTION 3: Embedding as []byte
// -----------------------------------------------------------------------
// Use []byte when you need raw file contents (e.g., private keys, certs, images)
//
//   //go:embed certs/server.crt
//   var serverCert []byte

// -----------------------------------------------------------------------
// SECTION 4: Embedding a Directory with embed.FS
// -----------------------------------------------------------------------
// embed.FS is a read-only virtual filesystem.
// It implements fs.FS, so it works with all fs-aware functions.
//
//   //go:embed templates/*
//   var templateFS embed.FS
//
//   //go:embed static
//   var staticFiles embed.FS   ← includes the "static" directory itself

// Simulate an embedded FS for demonstration
// In real code this replaces runtime file reads with embedded data:
//
//   //go:embed migrations/*.sql
//   var migrationsFS embed.FS

func embeddedFSDemo() {
	fmt.Println("embed.FS patterns (simulated — real usage needs actual files):")

	// How you'd use a real embed.FS:
	usage := `
  // In your Go file:
  //go:embed templates/*
  var templateFS embed.FS

  // Read a specific file:
  data, err := templateFS.ReadFile("templates/index.html")

  // Walk all files:
  fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
      fmt.Println(path)
      return nil
  })

  // Use with http.FileServer:
  sub, _ := fs.Sub(templateFS, "static")
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

  // Use with html/template:
  tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
`
	fmt.Println(usage)
}

// -----------------------------------------------------------------------
// SECTION 5: Using embed.FS (in-memory simulation)
// -----------------------------------------------------------------------
// Since we can't embed real files in this standalone demo,
// we'll show the SAME API using a custom fs.FS backed by a map.

// mockFS simulates an embed.FS for demonstration purposes
type mockFS map[string]string

func (m mockFS) Open(name string) (fs.File, error) {
	return nil, fmt.Errorf("mockFS.Open: not implemented in demo")
}

func (m mockFS) ReadFile(name string) ([]byte, error) {
	if content, ok := m[name]; ok {
		return []byte(content), nil
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func simulatedEmbedUsage() {
	fmt.Println("Simulated embed.FS usage:")

	// This represents what would be embedded from disk at compile time
	files := mockFS{
		"templates/index.html":  "<html><body>Hello, {{.Name}}!</body></html>",
		"templates/error.html":  "<html><body>Error: {{.Error}}</body></html>",
		"config/default.yaml":   "host: localhost\nport: 8080\n",
		"migrations/001_init.sql": "CREATE TABLE users (id INT PRIMARY KEY, name TEXT);",
		"migrations/002_add_email.sql": "ALTER TABLE users ADD COLUMN email TEXT;",
	}

	// Read a single file
	data, err := files.ReadFile("config/default.yaml")
	if err == nil {
		fmt.Printf("  config: %s\n", strings.TrimSpace(string(data)))
	}

	// List all files (in real embed.FS: fs.WalkDir)
	fmt.Println("  embedded files:")
	for name := range files {
		fmt.Printf("    %s\n", name)
	}
}

// -----------------------------------------------------------------------
// SECTION 6: Common Use Cases
// -----------------------------------------------------------------------

func useCases() {
	fmt.Println("\nCommon //go:embed use cases:")

	cases := []struct{ useCase, example string }{
		{
			"HTML templates",
			"//go:embed templates/*\nvar templateFS embed.FS",
		},
		{
			"Static web assets (CSS, JS, images)",
			"//go:embed static\nvar staticFS embed.FS\n// serve: http.FileServer(http.FS(staticFS))",
		},
		{
			"Database migrations",
			"//go:embed migrations/*.sql\nvar migrationsFS embed.FS",
		},
		{
			"Default config file",
			"//go:embed config/default.yaml\nvar defaultConfig []byte",
		},
		{
			"TLS certificates",
			"//go:embed certs/server.crt certs/server.key\nvar certFS embed.FS",
		},
		{
			"Version/build info",
			"//go:embed version.txt\nvar version string",
		},
		{
			"Swagger/OpenAPI spec",
			"//go:embed api/swagger.json\nvar swaggerSpec []byte",
		},
	}

	for _, c := range cases {
		fmt.Printf("  USE CASE: %s\n", c.useCase)
		for _, line := range strings.Split(c.example, "\n") {
			fmt.Printf("    %s\n", line)
		}
		fmt.Println()
	}
}

// -----------------------------------------------------------------------
// SECTION 7: Glob Patterns in //go:embed
// -----------------------------------------------------------------------

func embedPatterns() {
	fmt.Println("//go:embed glob patterns:")
	patterns := []struct{ pattern, desc string }{
		{"//go:embed file.txt", "embed a single file"},
		{"//go:embed dir", "embed entire directory (all files, no hidden)"},
		{"//go:embed dir/*", "embed all non-hidden files in dir"},
		{"//go:embed dir/**", "embed all files including subdirectories"},
		{"//go:embed *.html", "embed all .html files in current dir"},
		{"//go:embed a.txt b.txt", "embed multiple specific files"},
		{"//go:embed templates static", "embed two directories into one embed.FS"},
	}
	for _, p := range patterns {
		fmt.Printf("  %-40s → %s\n", p.pattern, p.desc)
	}

	fmt.Println("\nNotes:")
	fmt.Println("  - Hidden files (.dotfiles) are excluded by default")
	fmt.Println("  - Use 'all:dir' pattern to include hidden files: //go:embed all:dir")
	fmt.Println("  - Only works with package-level variables (not local vars)")
	fmt.Println("  - Paths must be within the module")
}

// -----------------------------------------------------------------------
// SECTION 8: embed.FS vs os package
// -----------------------------------------------------------------------

func embedVsOS() {
	fmt.Println("\nembed.FS vs os package:")
	fmt.Println("  os.ReadFile:")
	fmt.Println("    + Reads from disk at RUNTIME")
	fmt.Println("    + Files can be changed without recompiling")
	fmt.Println("    - Requires files to exist at correct path when deployed")
	fmt.Println("    - Deployment must include all asset files")
	fmt.Println()
	fmt.Println("  embed.FS:")
	fmt.Println("    + Files baked INTO the binary at COMPILE TIME")
	fmt.Println("    + Single binary deployment — no missing files")
	fmt.Println("    + Faster reads (from memory, not disk)")
	fmt.Println("    - Binary size increases")
	fmt.Println("    - Must recompile to update files")
}

// Declare a package-level embed.FS even though no directive is used
// This is just so the embed import is used
var _ embed.FS

func main() {
	fmt.Printf("Version (embedded): %s\n\n", version)
	embeddedFSDemo()
	simulatedEmbedUsage()
	useCases()
	embedPatterns()
	embedVsOS()
}
