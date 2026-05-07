package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
}
