// Package main demonstrates os/exec — running external commands from Go.
// Topics: Run, Output, stdin/stdout pipes, streaming, context cancellation.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Running a Simple Command
// -----------------------------------------------------------------------
// exec.Command(name, args...) creates a Cmd.
// The command is NOT run until you call Run(), Output(), or Start().

func simpleCommand() {
	fmt.Println("Simple command:")

	// echo "hello from exec"
	cmd := exec.Command("echo", "hello from exec")

	// Output() runs the command and returns stdout as []byte
	// Also returns an error if exit code != 0
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	fmt.Printf("  output: %q\n", strings.TrimSpace(string(output)))
}

// -----------------------------------------------------------------------
// SECTION 2: Capturing stdout and stderr separately
// -----------------------------------------------------------------------

func captureOutput() {
	fmt.Println("\nCapturing stdout + stderr:")

	cmd := exec.Command("sh", "-c", "echo stdout; echo stderr >&2; exit 0")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	fmt.Printf("  stdout: %q\n", strings.TrimSpace(stdout.String()))
	fmt.Printf("  stderr: %q\n", strings.TrimSpace(stderr.String()))
	fmt.Printf("  err:    %v\n", err)
}

// -----------------------------------------------------------------------
// SECTION 3: CombinedOutput — stdout + stderr merged
// -----------------------------------------------------------------------

func combinedOutput() {
	fmt.Println("\nCombinedOutput:")

	cmd := exec.Command("sh", "-c", "echo line1; echo line2 >&2; echo line3")
	out, err := cmd.CombinedOutput()
	fmt.Printf("  combined: %q\n", strings.TrimSpace(string(out)))
	fmt.Printf("  err:      %v\n", err)
}

// -----------------------------------------------------------------------
// SECTION 4: Exit Code and exec.ExitError
// -----------------------------------------------------------------------
// When a command exits with non-zero status, Run/Output returns *exec.ExitError.
// ExitError.ExitCode() gives you the exit code.

func exitCode() {
	fmt.Println("\nExit code handling:")

	cmd := exec.Command("sh", "-c", "exit 42")
	err := cmd.Run()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("  exit code: %d\n", exitErr.ExitCode())
			fmt.Printf("  stderr: %q\n", string(exitErr.Stderr))
		} else {
			fmt.Printf("  other error: %v\n", err)
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 5: Providing stdin
// -----------------------------------------------------------------------

func stdinInput() {
	fmt.Println("\nProviding stdin:")

	cmd := exec.Command("cat") // cat echoes stdin to stdout
	cmd.Stdin = strings.NewReader("line1\nline2\nline3\n")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	fmt.Printf("  output: %q\n", strings.TrimSpace(string(output)))
}

// -----------------------------------------------------------------------
// SECTION 6: Streaming Output (Pipe)
// -----------------------------------------------------------------------
// For long-running commands, use StdoutPipe() to stream output
// instead of waiting for it all to accumulate in memory.

func streamingOutput() {
	fmt.Println("\nStreaming output with pipe:")

	// Simulate a slow command that outputs line by line
	cmd := exec.Command("sh", "-c", `
		for i in 1 2 3; do
			echo "line $i"
			sleep 0.05
		done
	`)

	// Get a pipe to stdout BEFORE calling Start()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("  pipe error: %v\n", err)
		return
	}

	// Start the command (non-blocking)
	if err := cmd.Start(); err != nil {
		fmt.Printf("  start error: %v\n", err)
		return
	}

	// Stream output line by line
	buf := make([]byte, 256)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			fmt.Printf("  streamed: %s", buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("  read error: %v\n", err)
			break
		}
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Printf("  wait error: %v\n", err)
	}
}

// -----------------------------------------------------------------------
// SECTION 7: Context Cancellation
// -----------------------------------------------------------------------
// Use CommandContext to automatically kill the command when context is cancelled.
// Essential for: request-scoped commands, timeout-bound operations.

func contextCancellation() {
	fmt.Println("\nContext cancellation:")

	// 100ms timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This command takes 5 seconds — will be killed by context timeout
	cmd := exec.CommandContext(ctx, "sleep", "5")

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		if ctx.Err() != nil {
			fmt.Printf("  command killed after %v (context: %v)\n", elapsed, ctx.Err())
		} else {
			fmt.Printf("  error: %v\n", err)
		}
	}
}

// -----------------------------------------------------------------------
// SECTION 8: Environment Variables
// -----------------------------------------------------------------------

func envVariables() {
	fmt.Println("\nEnvironment variables:")

	cmd := exec.Command("sh", "-c", "echo $MY_VAR and $OTHER_VAR")

	// Option 1: inherit all env vars from parent process
	// cmd.Env = os.Environ()

	// Option 2: set specific env vars (use os.Environ() to inherit parent)
	cmd.Env = append(os.Environ(),
		"MY_VAR=hello",
		"OTHER_VAR=world",
	)

	output, _ := cmd.Output()
	fmt.Printf("  output: %q\n", strings.TrimSpace(string(output)))
}

// -----------------------------------------------------------------------
// SECTION 9: Working Directory
// -----------------------------------------------------------------------

func workingDirectory() {
	fmt.Println("\nWorking directory:")

	cmd := exec.Command("pwd")
	cmd.Dir = "/tmp" // set working directory

	output, _ := cmd.Output()
	fmt.Printf("  pwd from /tmp: %q\n", strings.TrimSpace(string(output)))
}

// -----------------------------------------------------------------------
// SECTION 10: LookPath — Find Executable
// -----------------------------------------------------------------------

func lookPath() {
	fmt.Println("\nLookPath:")

	// Find the full path of an executable
	path, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("  sh not found: %v\n", err)
	} else {
		fmt.Printf("  sh path: %s\n", path)
	}

	// Check if a command exists before running
	_, err = exec.LookPath("nonexistent_command_xyz")
	if err != nil {
		fmt.Println("  nonexistent_command not found (expected)")
	}
}

// -----------------------------------------------------------------------
// SECTION 11: Security — Avoid Shell Injection
// -----------------------------------------------------------------------

func securityNotes() {
	fmt.Println("\nSecurity — avoid shell injection:")
	fmt.Println(`
  BAD: Building shell commands with user input
    userInput := "file.txt; rm -rf /"
    cmd := exec.Command("sh", "-c", "cat " + userInput)  // INJECTION!

  GOOD: Pass arguments directly (no shell interpretation)
    cmd := exec.Command("cat", userInput)  // safe — no shell

  ALSO BAD: Using exec.Command("sh", "-c", fmt.Sprintf("...%s...", input))

  RULE: Never use "sh -c" with user-controlled input.
        Pass arguments as separate strings to exec.Command.
`)
}

func main() {
	simpleCommand()
	captureOutput()
	combinedOutput()
	exitCode()
	stdinInput()
	streamingOutput()
	contextCancellation()
	envVariables()
	workingDirectory()
	lookPath()
	securityNotes()
}
