// Package main explains Go modules, packages, and project structure.
// This file focuses on concepts — run "go mod init", "go get", etc. in terminal.
package main

import "fmt"

// -----------------------------------------------------------------------
// SECTION 1: Module Basics
// -----------------------------------------------------------------------
// A MODULE is a collection of packages with a go.mod file at its root.
// go.mod defines:
//   - module path (import path prefix)
//   - Go version
//   - Dependencies (require, replace, exclude)
//
// go.sum contains cryptographic hashes of dependencies for verification.
//
// Key commands:
//   go mod init <module-path>   — create go.mod
//   go get <pkg>@<version>      — add/update a dependency
//   go mod tidy                 — remove unused deps, add missing ones
//   go mod vendor               — copy deps to ./vendor
//   go list -m all              — list all dependencies

// -----------------------------------------------------------------------
// SECTION 2: Package Organization
// -----------------------------------------------------------------------
// Go conventions:
//
//   myapp/
//   ├── go.mod
//   ├── main.go            ← package main (entry point)
//   ├── internal/          ← only importable by code in parent dir tree
//   │   └── auth/
//   │       └── auth.go    ← package auth
//   ├── pkg/               ← public reusable packages (community convention)
//   │   └── httpclient/
//   │       └── client.go
//   └── cmd/               ← multiple binaries
//       ├── server/
//       │   └── main.go    ← package main
//       └── migrate/
//           └── main.go    ← package main
//
// Rules:
//   1. One package per directory (except _test suffix)
//   2. Package name = directory name (usually)
//   3. Exported = starts with uppercase
//   4. Unexported = starts with lowercase (package-private)

// -----------------------------------------------------------------------
// SECTION 3: init() Function
// -----------------------------------------------------------------------
// init() runs automatically before main(), in the order:
//   1. Package-level variables are initialized
//   2. init() functions run (can be multiple per file, multiple per package)
//   3. main() runs
//
// Use init() for:
//   - Registering drivers, codecs, plugins
//   - Validating configuration
//   - One-time setup that can't fail gracefully
//
// Avoid overusing init() — it's hard to test and control.

var configValue = initConfig() // package-level var initialized first

func initConfig() string {
	fmt.Println("  [init] package var initialized")
	return "default-config"
}

func init() {
	// First init() in this file
	fmt.Println("  [init] first init() called")
}

func init() {
	// You CAN have multiple init() functions per file
	fmt.Println("  [init] second init() called")
}

// -----------------------------------------------------------------------
// SECTION 4: Blank Import (Side Effects)
// -----------------------------------------------------------------------
// Import with _ to run a package's init() without using any exported symbols.
// Common use: registering drivers, image formats, etc.
//
// Example (don't run — just for illustration):
//   import _ "github.com/lib/pq"          // registers postgres driver
//   import _ "image/png"                  // registers PNG decoder
//   import _ "net/http/pprof"             // registers /debug/pprof handlers

// -----------------------------------------------------------------------
// SECTION 5: Internal Packages
// -----------------------------------------------------------------------
// The "internal" directory is enforced by the Go toolchain.
// Only packages rooted at the parent of "internal" can import it.
//
//   github.com/you/app/internal/auth   ← can be imported by:
//     github.com/you/app/...           ← yes
//     github.com/other/pkg/...         ← NO — compile error
//
// Use internal/ for implementation details you don't want to expose.

// -----------------------------------------------------------------------
// SECTION 6: Build Tags
// -----------------------------------------------------------------------
// Build tags include/exclude files from compilation.
// Syntax (Go 1.17+):
//   //go:build linux
//   //go:build !windows
//   //go:build integration
//
// Run with: go test -tags integration ./...
//
// Common uses:
//   - OS-specific code
//   - Integration tests that need external services
//   - Feature flags during development

// -----------------------------------------------------------------------
// SECTION 7: go:generate
// -----------------------------------------------------------------------
// go generate runs commands before building. Useful for:
//   - Generating mocks (mockgen)
//   - Generating protobuf code (protoc)
//   - Embedding files (before Go 1.16 embed)
//
// Example:
//   //go:generate mockgen -source=store.go -destination=mock_store.go
//
// Run with: go generate ./...

func main() {
	fmt.Println("Module and Package concepts:")
	fmt.Println("  (init functions already ran above)")
	fmt.Printf("  configValue = %q\n", configValue)

	fmt.Println("\nKey go.mod commands to know:")
	commands := []struct{ cmd, desc string }{
		{"go mod init myapp", "create new module"},
		{"go get pkg@v1.2.3", "add/update dependency"},
		{"go mod tidy", "clean up go.mod and go.sum"},
		{"go mod download", "download modules to cache"},
		{"go list -m all", "list all dependencies"},
		{"go mod why pkg", "why is pkg a dependency"},
		{"go build ./...", "build all packages"},
		{"go test ./...", "test all packages"},
		{"go vet ./...", "report likely mistakes"},
	}
	for _, c := range commands {
		fmt.Printf("  %-35s ← %s\n", c.cmd, c.desc)
	}
}
