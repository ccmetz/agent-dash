package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesBuiltInDefaultsWhenLocalConfigIsMissing(t *testing.T) {
	t.Setenv("AGENT_DASH_CONFIG", filepath.Join(t.TempDir(), "missing.json"))

	config, err := Load()
	if err != nil {
		t.Fatalf("expected missing config to use defaults: %v", err)
	}

	if config.AnalyticsStorePath != filepath.Join("data", "agent-dash.sqlite") {
		t.Fatalf("expected default Analytics Store path, got %q", config.AnalyticsStorePath)
	}
}

func TestLoadUsesValuesFromJSONConfigFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "local.json")
	if err := os.WriteFile(configPath, []byte(`{"analyticsStorePath":"custom/analytics.sqlite"}`), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	t.Setenv("AGENT_DASH_CONFIG", configPath)

	config, err := Load()
	if err != nil {
		t.Fatalf("expected config file to load: %v", err)
	}

	if config.AnalyticsStorePath != "custom/analytics.sqlite" {
		t.Fatalf("expected configured Analytics Store path, got %q", config.AnalyticsStorePath)
	}
}

func TestLoadReturnsErrorForInvalidJSONConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "local.json")
	if err := os.WriteFile(configPath, []byte(`{"analyticsStorePath":`), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	t.Setenv("AGENT_DASH_CONFIG", configPath)

	if _, err := Load(); err == nil {
		t.Fatalf("expected invalid config JSON to return an error")
	}
}
