package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var appLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

func ConfigureLogger(cfg AppConfig) error {
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.LogLevel))
	if err != nil {
		return fmt.Errorf("invalid log_level %q: %w", cfg.LogLevel, err)
	}

	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	if strings.EqualFold(cfg.LogFormat, "json") {
		appLogger = zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()
		return nil
	}
	if !strings.EqualFold(cfg.LogFormat, "console") {
		return fmt.Errorf("invalid log_format %q: expected console or json", cfg.LogFormat)
	}

	appLogger = zerolog.New(output).Level(level).With().Timestamp().Logger()
	return nil
}
