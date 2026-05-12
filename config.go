package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	SavePath  string `yaml:"save_path"`
	ChunkPath string `yaml:"chunk_path"`
	LogLevel  string `yaml:"log_level"`
	LogFormat string `yaml:"log_format"`
}

type RuntimeOptions struct {
	CLI        bool
	ConfigPath string
}

func ParseRuntimeOptions(args []string) (RuntimeOptions, error) {
	options := RuntimeOptions{}
	flags := flag.NewFlagSet("rainstorm-peer", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&options.CLI, "cli", false, "run the interactive CLI instead of the GUI")
	flags.StringVar(&options.ConfigPath, "config", "", "path to a YAML configuration file")

	if len(args) == 0 {
		return options, nil
	}
	if err := flags.Parse(args[1:]); err != nil {
		return RuntimeOptions{}, err
	}
	return options, nil
}

func LoadAppConfig(configPath string, baseDir string) (AppConfig, error) {
	if baseDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return AppConfig{}, err
		}
		baseDir = wd
	}

	cfg := AppConfig{}
	if configPath != "" {
		content, err := os.ReadFile(configPath)
		if err != nil {
			return AppConfig{}, fmt.Errorf("read config %q: %w", configPath, err)
		}
		if err := yaml.Unmarshal(content, &cfg); err != nil {
			return AppConfig{}, fmt.Errorf("parse config %q: %w", configPath, err)
		}
	}

	return applyConfigDefaults(cfg, baseDir), nil
}

func applyConfigDefaults(cfg AppConfig, baseDir string) AppConfig {
	if cfg.SavePath == "" {
		cfg.SavePath = os.Getenv("RSTM_SAVE_PATH")
	}
	if cfg.SavePath == "" {
		cfg.SavePath = filepath.Join(baseDir, "rstm_save")
	}
	if cfg.ChunkPath == "" {
		cfg.ChunkPath = filepath.Join(cfg.SavePath, "chunk_path")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogFormat == "" {
		cfg.LogFormat = "console"
	}

	cfg.LogLevel = strings.ToLower(cfg.LogLevel)
	cfg.LogFormat = strings.ToLower(cfg.LogFormat)
	return cfg
}
