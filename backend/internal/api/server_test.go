package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStatusUsesBuiltInConfigDefaultsWithoutLocalConfig(t *testing.T) {
	t.Setenv("AGENT_DASH_CONFIG", filepath.Join(t.TempDir(), "missing.json"))

	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	response := httptest.NewRecorder()

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var body struct {
		OK                 bool   `json:"ok"`
		AnalyticsStorePath string `json:"analyticsStorePath"`
		UsageSource        struct {
			Name      string `json:"name"`
			Path      string `json:"path"`
			Available bool   `json:"available"`
			State     string `json:"state"`
		} `json:"usageSource"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if !body.OK {
		t.Fatalf("expected status to be ok")
	}
	if body.AnalyticsStorePath != filepath.Join("data", "agent-dash.sqlite") {
		t.Fatalf("expected default Analytics Store path, got %q", body.AnalyticsStorePath)
	}
	if body.UsageSource.Name != "OpenCode" {
		t.Fatalf("expected OpenCode Usage Source, got %q", body.UsageSource.Name)
	}
}

func TestStatusReportsConfiguredOpenCodeUsageSourceAvailable(t *testing.T) {
	usageSourcePath := filepath.Join(t.TempDir(), "opencode.db")
	if err := os.WriteFile(usageSourcePath, []byte("sqlite"), 0o600); err != nil {
		t.Fatalf("failed to write Usage Source database: %v", err)
	}
	configPath := filepath.Join(t.TempDir(), "local.json")
	if err := os.WriteFile(configPath, []byte(`{"openCodeDatabasePath":"`+usageSourcePath+`"}`), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	t.Setenv("AGENT_DASH_CONFIG", configPath)

	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	response := httptest.NewRecorder()

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var body struct {
		UsageSource struct {
			Path      string `json:"path"`
			Available bool   `json:"available"`
			State     string `json:"state"`
		} `json:"usageSource"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if body.UsageSource.Path != usageSourcePath {
		t.Fatalf("expected configured Usage Source path, got %q", body.UsageSource.Path)
	}
	if !body.UsageSource.Available {
		t.Fatalf("expected Usage Source to be available")
	}
	if body.UsageSource.State != "available" {
		t.Fatalf("expected available Usage Source state, got %q", body.UsageSource.State)
	}
}

func TestStatusReportsMissingOpenCodeUsageSourceWithoutFailing(t *testing.T) {
	usageSourcePath := filepath.Join(t.TempDir(), "missing-opencode.db")
	configPath := filepath.Join(t.TempDir(), "local.json")
	if err := os.WriteFile(configPath, []byte(`{"openCodeDatabasePath":"`+usageSourcePath+`"}`), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	t.Setenv("AGENT_DASH_CONFIG", configPath)

	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	response := httptest.NewRecorder()

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected missing Usage Source to be non-fatal, got status %d", response.Code)
	}

	var body struct {
		OK          bool `json:"ok"`
		UsageSource struct {
			Path      string `json:"path"`
			Available bool   `json:"available"`
			State     string `json:"state"`
		} `json:"usageSource"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if !body.OK {
		t.Fatalf("expected status to stay ok when Usage Source is missing")
	}
	if body.UsageSource.Path != usageSourcePath {
		t.Fatalf("expected configured Usage Source path, got %q", body.UsageSource.Path)
	}
	if body.UsageSource.Available {
		t.Fatalf("expected Usage Source to be unavailable")
	}
	if body.UsageSource.State != "missing" {
		t.Fatalf("expected missing Usage Source state, got %q", body.UsageSource.State)
	}
}
