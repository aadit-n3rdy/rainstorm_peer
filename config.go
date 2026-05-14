package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	SavePath  string `yaml:"save_path"`
	ChunkPath string `yaml:"chunk_path"`
	LogLevel  string `yaml:"log_level"`
	LogFormat string `yaml:"log_format"`
}

type RuntimeOptions struct {
	Config AppConfig
	CLI    bool
}

func DefaultAppConfig(getenv func(string) string, wd string) AppConfig {
	savePath := getenv("RSTM_SAVE_PATH")
	if savePath == "" {
		savePath = filepath.Join(wd, "rstm_save")
	}

	return AppConfig{
		SavePath:  savePath,
		ChunkPath: filepath.Join(savePath, "chunk_path"),
		LogLevel:  "info",
		LogFormat: "console",
	}
}

func LoadAppConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, err
	}
	return cfg, nil
}

func mergeAppConfig(base AppConfig, override AppConfig) AppConfig {
	chunkPathWasDefault := base.ChunkPath == filepath.Join(base.SavePath, "chunk_path")
	if override.SavePath != "" {
		base.SavePath = override.SavePath
	}
	if override.ChunkPath != "" {
		base.ChunkPath = override.ChunkPath
	}
	if override.LogLevel != "" {
		base.LogLevel = override.LogLevel
	}
	if override.LogFormat != "" {
		base.LogFormat = override.LogFormat
	}
	if base.ChunkPath == "" || (override.SavePath != "" && override.ChunkPath == "" && chunkPathWasDefault) {
		base.ChunkPath = filepath.Join(base.SavePath, "chunk_path")
	}
	return base
}

func ParseRuntimeOptions(args []string, getenv func(string) string, wd string) (RuntimeOptions, error) {
	fs := flag.NewFlagSet("rainstorm_peer", flag.ContinueOnError)

	var configPath string
	var opts RuntimeOptions
	opts.Config = DefaultAppConfig(getenv, wd)
	fs.BoolVar(&opts.CLI, "cli", false, "run the interactive CLI instead of the GUI")
	fs.StringVar(&configPath, "config", "", "path to a YAML configuration file")

	if err := fs.Parse(args); err != nil {
		return RuntimeOptions{}, err
	}

	if configPath != "" {
		cfg, err := LoadAppConfig(configPath)
		if err != nil {
			return RuntimeOptions{}, fmt.Errorf("load config %q: %w", configPath, err)
		}
		opts.Config = mergeAppConfig(opts.Config, cfg)
	}

	return opts, nil
}
