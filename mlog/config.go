package mlog

import (
	"io"
	"log/slog"
	"os"
)

type LogFormat string

const (
	TextFormat LogFormat = "text"
	JSONFormat LogFormat = "json"
)

type Config struct {
	Level slog.Level
	Format LogFormat
	Output io.Writer
	AddSource bool
}

func DefaultConfig() Config {
	return Config{
		Level:     slog.LevelInfo,
		Format:    TextFormat,
		Output:    os.Stdout,
		AddSource: false,
	}
}

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

func ConfigureDebug() {
	cfg := DefaultConfig()
	cfg.Level = slog.LevelDebug
	Configure(cfg)
}
