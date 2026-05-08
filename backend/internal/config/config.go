package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	AnalyticsStorePath   string `json:"analyticsStorePath"`
	OpenCodeDatabasePath string `json:"openCodeDatabasePath"`
}

func Default() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	return Config{
		AnalyticsStorePath:   filepath.Join("data", "agent-dash.sqlite"),
		OpenCodeDatabasePath: filepath.Join(homeDir, ".local", "share", "opencode", "opencode.db"),
	}
}

func Load() (Config, error) {
	configPath := os.Getenv("AGENT_DASH_CONFIG")
	if configPath == "" {
		configPath = filepath.Join("config", "local.json")
	}

	file, err := os.Open(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	config := Default()
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
