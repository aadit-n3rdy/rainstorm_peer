package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var appLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

func NewAppLogger(cfg AppConfig, output io.Writer) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("parse log level %q: %w", cfg.LogLevel, err)
	}

	writer := output
	switch cfg.LogFormat {
	case "console":
		writer = zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339}
	case "json":
	default:
		return zerolog.Logger{}, fmt.Errorf("unsupported log format %q", cfg.LogFormat)
	}

	return zerolog.New(writer).Level(level).With().Timestamp().Logger(), nil
}

func SetAppLogger(logger zerolog.Logger) {
	appLogger = logger
}
