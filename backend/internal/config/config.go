package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	AnalyticsStorePath string `json:"analyticsStorePath"`
}

func Default() Config {
	return Config{AnalyticsStorePath: filepath.Join("data", "agent-dash.sqlite")}
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
