package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRuntimeOptionsDefaults(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"-cli"}, func(string) string { return "" }, "/tmp/work")
	if err != nil {
		t.Fatalf("ParseRuntimeOptions returned error: %v", err)
	}

	if !opts.CLI {
		t.Fatal("expected -cli to enable CLI mode")
	}
	if opts.Config.SavePath != filepath.Join("/tmp/work", "rstm_save") {
		t.Fatalf("unexpected save path: %q", opts.Config.SavePath)
	}
	if opts.Config.ChunkPath != filepath.Join("/tmp/work", "rstm_save", "chunk_path") {
		t.Fatalf("unexpected chunk path: %q", opts.Config.ChunkPath)
	}
	if opts.Config.LogLevel != "info" {
		t.Fatalf("unexpected log level: %q", opts.Config.LogLevel)
	}
	if opts.Config.LogFormat != "console" {
		t.Fatalf("unexpected log format: %q", opts.Config.LogFormat)
	}
}

func TestParseRuntimeOptionsUsesEnvironmentSavePath(t *testing.T) {
	opts, err := ParseRuntimeOptions(nil, func(key string) string {
		if key == "RSTM_SAVE_PATH" {
			return "/var/lib/rainstorm"
		}
		return ""
	}, "/tmp/work")
	if err != nil {
		t.Fatalf("ParseRuntimeOptions returned error: %v", err)
	}

	if opts.Config.SavePath != "/var/lib/rainstorm" {
		t.Fatalf("unexpected save path: %q", opts.Config.SavePath)
	}
	if opts.Config.ChunkPath != filepath.Join("/var/lib/rainstorm", "chunk_path") {
		t.Fatalf("unexpected chunk path: %q", opts.Config.ChunkPath)
	}
}

func TestParseRuntimeOptionsMergesConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "rainstorm.yml")
	err := os.WriteFile(configPath, []byte(`
save_path: /srv/rainstorm/state
chunk_path: /srv/rainstorm/chunks
log_level: debug
log_format: json
`), 0o600)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	opts, err := ParseRuntimeOptions([]string{"-config", configPath}, func(string) string { return "" }, "/tmp/work")
	if err != nil {
		t.Fatalf("ParseRuntimeOptions returned error: %v", err)
	}

	if opts.Config.SavePath != "/srv/rainstorm/state" {
		t.Fatalf("unexpected save path: %q", opts.Config.SavePath)
	}
	if opts.Config.ChunkPath != "/srv/rainstorm/chunks" {
		t.Fatalf("unexpected chunk path: %q", opts.Config.ChunkPath)
	}
	if opts.Config.LogLevel != "debug" {
		t.Fatalf("unexpected log level: %q", opts.Config.LogLevel)
	}
	if opts.Config.LogFormat != "json" {
		t.Fatalf("unexpected log format: %q", opts.Config.LogFormat)
	}
}

func TestParseRuntimeOptionsMovesDefaultChunkPathWithConfiguredSavePath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "rainstorm.yml")
	err := os.WriteFile(configPath, []byte(`
save_path: /srv/rainstorm/state
`), 0o600)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	opts, err := ParseRuntimeOptions([]string{"-config", configPath}, func(string) string { return "" }, "/tmp/work")
	if err != nil {
		t.Fatalf("ParseRuntimeOptions returned error: %v", err)
	}

	if opts.Config.SavePath != "/srv/rainstorm/state" {
		t.Fatalf("unexpected save path: %q", opts.Config.SavePath)
	}
	if opts.Config.ChunkPath != filepath.Join("/srv/rainstorm/state", "chunk_path") {
		t.Fatalf("unexpected chunk path: %q", opts.Config.ChunkPath)
	}
}
