package mlog

import (
	"io"
	"log/slog"
	"os"
)

// LogFormat defines the output format for logs.
type LogFormat string

const (
	// TextFormat indicates logs should be formatted as human-readable text.
	TextFormat LogFormat = "text"
	// JSONFormat indicates logs should be formatted as JSON.
	JSONFormat LogFormat = "json"
)

// Config holds the configuration options for the logger.
type Config struct {
	// Level is the minimum log level that will be output.
	Level slog.Level
	// Format specifies whether logs should be in text or JSON format.
	Format LogFormat
	// Output is the destination for logs (e.g., os.Stdout).
	Output io.Writer
	// AddSource indicates whether to include source file and line information.
	AddSource bool
}

// DefaultConfig returns a default configuration for the logger.
// By default, logs are written to stdout at INFO level in text format.
func DefaultConfig() Config {
	return Config{
		Level:     slog.LevelInfo,
		Format:    TextFormat,
		Output:    os.Stdout,
		AddSource: false,
	}
}

// Configure sets up the global logger with the specified configuration.
func Configure(cfg Config) {
	var handler slog.Handler
	
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}
	
	switch cfg.Format {
	case JSONFormat:
		handler = slog.NewJSONHandler(cfg.Output, opts)
	default:
		handler = slog.NewTextHandler(cfg.Output, opts)
	}
	
	contextHandler := NewContextAwareHandler(handler)
	
	SetDefaultLogger(slog.New(contextHandler))
}

// ConfigureDebug is a convenience function that configures the logger for debug output.
// It uses the default configuration but sets the log level to DEBUG.
func ConfigureDebug() {
	cfg := DefaultConfig()
	cfg.Level = slog.LevelDebug
	Configure(cfg)
}
