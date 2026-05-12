package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAppConfigReadsYAMLAndAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "peer.yaml")
	savePath := filepath.Join(dir, "state")
	chunkPath := filepath.Join(dir, "chunks")

	content := []byte(strings.Join([]string{
		"save_path: " + savePath,
		"chunk_path: " + chunkPath,
		"log_level: debug",
		"log_format: json",
	}, "\n"))
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadAppConfig(configPath, dir)
	if err != nil {
		t.Fatalf("LoadAppConfig returned error: %v", err)
	}

	if cfg.SavePath != savePath {
		t.Fatalf("SavePath = %q, want %q", cfg.SavePath, savePath)
	}
	if cfg.ChunkPath != chunkPath {
		t.Fatalf("ChunkPath = %q, want %q", cfg.ChunkPath, chunkPath)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Fatalf("LogFormat = %q, want json", cfg.LogFormat)
	}
}

func TestLoadAppConfigUsesDefaultsForMissingFilePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RSTM_SAVE_PATH", "")

	cfg, err := LoadAppConfig("", dir)
	if err != nil {
		t.Fatalf("LoadAppConfig returned error: %v", err)
	}

	wantSavePath := filepath.Join(dir, "rstm_save")
	wantChunkPath := filepath.Join(wantSavePath, "chunk_path")
	if cfg.SavePath != wantSavePath {
		t.Fatalf("SavePath = %q, want %q", cfg.SavePath, wantSavePath)
	}
	if cfg.ChunkPath != wantChunkPath {
		t.Fatalf("ChunkPath = %q, want %q", cfg.ChunkPath, wantChunkPath)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.LogFormat != "console" {
		t.Fatalf("LogFormat = %q, want console", cfg.LogFormat)
	}
}

func TestParseRuntimeOptionsAcceptsConfigForCLIAndGUI(t *testing.T) {
	cliOptions, err := ParseRuntimeOptions([]string{"rainstorm-peer", "-cli", "-config", "peer.yaml"})
	if err != nil {
		t.Fatalf("ParseRuntimeOptions for CLI returned error: %v", err)
	}
	if !cliOptions.CLI {
		t.Fatal("CLI option should be true when -cli is present")
	}
	if cliOptions.ConfigPath != "peer.yaml" {
		t.Fatalf("ConfigPath = %q, want peer.yaml", cliOptions.ConfigPath)
	}

	guiOptions, err := ParseRuntimeOptions([]string{"rainstorm-peer", "-config", "gui.yaml"})
	if err != nil {
		t.Fatalf("ParseRuntimeOptions for GUI returned error: %v", err)
	}
	if guiOptions.CLI {
		t.Fatal("CLI option should be false when -cli is absent")
	}
	if guiOptions.ConfigPath != "gui.yaml" {
		t.Fatalf("ConfigPath = %q, want gui.yaml", guiOptions.ConfigPath)
	}
}

func TestNewAppLoggerHonorsJSONFormatAndLevel(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewAppLogger(AppConfig{LogLevel: "debug", LogFormat: "json"}, &buf)
	if err != nil {
		t.Fatalf("NewAppLogger returned error: %v", err)
	}

	logger.Debug().Msg("debug-enabled")

	output := buf.String()
	if !strings.Contains(output, `"level":"debug"`) {
		t.Fatalf("logger output %q does not contain debug level", output)
	}
	if !strings.Contains(output, `"message":"debug-enabled"`) {
		t.Fatalf("logger output %q does not contain message", output)
	}
}
