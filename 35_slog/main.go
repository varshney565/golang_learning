// Package main demonstrates log/slog — structured logging (Go 1.21+).
// slog is the new standard for structured logging in Go.
// Topics: basic usage, levels, attributes, handlers, custom handler.
package main

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: What is slog?
// -----------------------------------------------------------------------
// log/slog (Go 1.21) is the official structured logging package.
// Before slog, everyone used different libraries (zap, logrus, zerolog).
// Now there's a standard interface, making libraries interoperable.
//
// Key concepts:
//   Level     — DEBUG, INFO, WARN, ERROR (and custom)
//   Record    — a single log entry with time, level, message, and attrs
//   Attr      — a key-value pair (typed for performance)
//   Handler   — where logs go (text, JSON, custom)
//   Logger    — combines handler + context

// -----------------------------------------------------------------------
// SECTION 2: Default Logger
// -----------------------------------------------------------------------

func defaultLogger() {
	slog.Info("Server starting", "host", "localhost", "port", 8080)
	slog.Debug("Debug message — won't show (default level is INFO)")
	slog.Warn("Disk space low", "used_percent", 85)
	slog.Error("Connection failed", "host", "db.example.com", "err", "timeout")
}

// -----------------------------------------------------------------------
// SECTION 3: Creating Loggers
// -----------------------------------------------------------------------

func createLoggers() {
	// Text handler — human readable
	textLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // show all levels
	}))

	textLogger.Info("text handler", "format", "key=value")
	textLogger.Debug("debug visible now", "reason", "level=DEBUG")

	// JSON handler — machine readable (use in production)
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// ReplaceAttr: customize attribute names/values
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Rename the time key from "time" to "ts"
			if a.Key == slog.TimeKey {
				return slog.Attr{Key: "ts", Value: a.Value}
			}
			// Remove the "source" key if present
			if a.Key == slog.SourceKey {
				return slog.Attr{} // empty attr = omit
			}
			return a
		},
	}))

	jsonLogger.Info("json handler", "format", "JSON")
	jsonLogger.Warn("something to watch", "threshold", 90)
}

// -----------------------------------------------------------------------
// SECTION 4: Typed Attributes (slog.Attr)
// -----------------------------------------------------------------------
// Key-value pairs can be passed as alternating args (any, any)
// OR as typed slog.Attr for zero-allocation performance.

func typedAttributes() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Shorthand: alternating key-value args (convenient, slight allocation)
	logger.Info("user logged in", "user_id", 42, "ip", "192.168.1.1")

	// Typed attrs: slog.String, slog.Int, slog.Bool, slog.Duration, slog.Any
	// More verbose but zero-allocation (no interface boxing)
	logger.Info("typed attrs",
		slog.String("method", "GET"),
		slog.String("path", "/api/users"),
		slog.Int("status", 200),
		slog.Duration("latency", 15*time.Millisecond),
		slog.Bool("cached", false),
	)

	// slog.Group — group related attrs under a namespace
	logger.Info("request",
		slog.Group("http",
			slog.String("method", "POST"),
			slog.Int("status", 201),
			slog.String("path", "/users"),
		),
		slog.Group("user",
			slog.Int("id", 42),
			slog.String("email", "alice@example.com"),
		),
	)
}

// -----------------------------------------------------------------------
// SECTION 5: Logger with Context — slog.With
// -----------------------------------------------------------------------
// slog.With creates a child logger with pre-attached attributes.
// Use this to add request-ID, user-ID, etc. once per request.

func withLogger() {
	base := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create a logger with persistent fields
	requestLogger := base.With(
		slog.String("request_id", "req-abc-123"),
		slog.String("user_id", "user-456"),
		slog.String("service", "api"),
	)

	// All subsequent logs include these fields automatically
	requestLogger.Info("handling request", slog.String("path", "/users"))
	requestLogger.Info("query executed", slog.Duration("duration", 5*time.Millisecond))
	requestLogger.Warn("rate limit approaching", slog.Int("remaining", 10))
}

// -----------------------------------------------------------------------
// SECTION 6: slog with context.Context
// -----------------------------------------------------------------------
// slog.InfoContext / slog.WarnContext etc. accept a context.
// This allows handlers to extract trace IDs or other context values.

func withContext() {
	ctx := context.Background()
	// In real code, you'd store trace_id in context:
	// ctx = context.WithValue(ctx, "trace_id", "trace-xyz")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Context-aware logging — handler can inspect context
	logger.InfoContext(ctx, "processing started", slog.String("step", "init"))
	logger.WarnContext(ctx, "slow query detected", slog.Duration("duration", 500*time.Millisecond))
}

// -----------------------------------------------------------------------
// SECTION 7: Custom slog.Handler
// -----------------------------------------------------------------------
// Implement slog.Handler to send logs anywhere: Datadog, Sentry, Slack, etc.
// Must implement: Enabled, Handle, WithAttrs, WithGroup.

type PrettyHandler struct {
	attrs []slog.Attr
	level slog.Level
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	// Custom format: emoji + level + message + attrs
	emoji := map[slog.Level]string{
		slog.LevelDebug: "🔍",
		slog.LevelInfo:  "ℹ️ ",
		slog.LevelWarn:  "⚠️ ",
		slog.LevelError: "❌",
	}

	icon := emoji[r.Level]
	if icon == "" {
		icon = "•"
	}

	// Print using fmt would normally write to os.Stderr/file
	// Using Println here for demo purposes
	line := r.Time.Format("15:04:05") + " " + icon + " " + r.Level.String() + " " + r.Message
	r.Attrs(func(a slog.Attr) bool {
		line += " " + a.Key + "=" + a.Value.String()
		return true
	})
	for _, a := range h.attrs {
		line += " " + a.Key + "=" + a.Value.String()
	}

	// Write to stdout (in production: write to file/network/etc.)
	_, err := os.Stdout.WriteString(line + "\n")
	return err
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &PrettyHandler{attrs: newAttrs, level: h.level}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	// Groups prefix all subsequent attr keys with "name."
	// Simplified implementation for demo
	return h
}

func customHandler() {
	logger := slog.New(&PrettyHandler{level: slog.LevelDebug})

	logger.Debug("starting up", slog.String("env", "dev"))
	logger.Info("server ready", slog.String("addr", ":8080"))
	logger.Warn("high memory", slog.Int("mb", 450))
	logger.Error("database error", slog.String("err", "connection refused"))
}

// -----------------------------------------------------------------------
// SECTION 8: Setting the Default Logger
// -----------------------------------------------------------------------

func setDefaultLogger() {
	// Override the package-level default logger
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true, // adds file:line to each log entry
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Now slog.Info, slog.Warn etc. use this logger
	slog.Info("default logger overridden", slog.String("format", "JSON to stderr"))
}

func main() {
	slog.Info("=== default logger ===")
	defaultLogger()

	slog.Info("\n=== creating loggers ===")
	createLoggers()

	slog.Info("\n=== typed attributes ===")
	typedAttributes()

	slog.Info("\n=== logger with context ===")
	withLogger()

	slog.Info("\n=== context-aware logging ===")
	withContext()

	slog.Info("\n=== custom handler ===")
	customHandler()

	slog.Info("\n=== setting default ===")
	setDefaultLogger()
}
