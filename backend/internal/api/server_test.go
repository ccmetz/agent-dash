package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
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
	createSupportedOpenCodeDatabase(t, usageSourcePath)
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

func TestStatusReportsUnsupportedOpenCodeUsageSourceSchemaWithoutFailing(t *testing.T) {
	usageSourcePath := filepath.Join(t.TempDir(), "opencode.db")
	db := openSQLite(t, usageSourcePath)
	defer db.Close()
	execSQL(t, db, `create table project (id text primary key)`)

	configPath := filepath.Join(t.TempDir(), "local.json")
	if err := os.WriteFile(configPath, []byte(`{"openCodeDatabasePath":"`+usageSourcePath+`"}`), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	t.Setenv("AGENT_DASH_CONFIG", configPath)

	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	response := httptest.NewRecorder()

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected unsupported Usage Source schema to be non-fatal, got status %d", response.Code)
	}

	var body struct {
		OK          bool `json:"ok"`
		UsageSource struct {
			Available bool   `json:"available"`
			State     string `json:"state"`
		} `json:"usageSource"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if !body.OK {
		t.Fatalf("expected status to stay ok when Usage Source schema is unsupported")
	}
	if body.UsageSource.Available {
		t.Fatalf("expected unsupported Usage Source schema to be unavailable")
	}
	if body.UsageSource.State != "unsupported_schema" {
		t.Fatalf("expected unsupported schema Usage Source state, got %q", body.UsageSource.State)
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

func TestManualUsageSyncEndpointSyncsOpenCodeUsageAndStatusShowsLastRun(t *testing.T) {
	dir := t.TempDir()
	usageSourcePath := filepath.Join(dir, "opencode.db")
	storePath := filepath.Join(dir, "analytics.db")
	createSupportedOpenCodeDatabase(t, usageSourcePath)
	db := openSQLite(t, usageSourcePath)
	execSQL(t, db, `insert into project (id, worktree, name, time_created, time_updated) values ('proj_1', '/work/agent-dash', 'Agent Dash', 1000, 2000)`)
	execSQL(t, db, `insert into session (id, project_id, title, time_created, time_updated, time_archived) values ('ses_1', 'proj_1', 'Manual sync', 3000, 4000, null)`)
	execSQL(t, db, `insert into message (id, session_id, time_created, time_updated, data) values ('msg_1', 'ses_1', 5000, 6000, '{"role":"assistant","modelID":"claude-opus-4-5","providerID":"opencode","cost":0.25,"tokens":{"input":10,"output":20,"reasoning":3,"cache":{"read":4,"write":5}},"finish":"stop"}')`)
	db.Close()
	configPath := filepath.Join(dir, "local.json")
	if err := os.WriteFile(configPath, []byte(`{"openCodeDatabasePath":"`+usageSourcePath+`","analyticsStorePath":"`+storePath+`"}`), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	t.Setenv("AGENT_DASH_CONFIG", configPath)

	request := httptest.NewRequest(http.MethodPost, "/api/usage-sync", nil)
	response := httptest.NewRecorder()
	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected manual sync status 200, got %d", response.Code)
	}
	var syncBody struct {
		Status   string `json:"status"`
		Inserted int    `json:"inserted"`
	}
	if err := json.NewDecoder(response.Body).Decode(&syncBody); err != nil {
		t.Fatalf("expected manual sync JSON response: %v", err)
	}
	if syncBody.Status != "success" || syncBody.Inserted != 3 {
		t.Fatalf("expected successful manual sync counts, got %+v", syncBody)
	}

	statusRequest := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	statusResponse := httptest.NewRecorder()
	NewServer().ServeHTTP(statusResponse, statusRequest)
	var statusBody struct {
		UsageSync struct {
			Status  string `json:"status"`
			LastRun *struct {
				Inserted int `json:"inserted"`
			} `json:"lastRun"`
			PollSeconds int    `json:"pollSeconds"`
			NextPollAt  string `json:"nextPollAt"`
		} `json:"usageSync"`
	}
	if err := json.NewDecoder(statusResponse.Body).Decode(&statusBody); err != nil {
		t.Fatalf("expected status JSON response: %v", err)
	}
	if statusBody.UsageSync.Status != "success" || statusBody.UsageSync.LastRun.Inserted != 3 || statusBody.UsageSync.PollSeconds != 60 || statusBody.UsageSync.NextPollAt == "" {
		t.Fatalf("expected status sync diagnostics, got %+v", statusBody.UsageSync)
	}
}

func createSupportedOpenCodeDatabase(t *testing.T, path string) {
	t.Helper()
	db := openSQLite(t, path)
	defer db.Close()
	execSQL(t, db, `create table project (id text primary key, worktree text not null, name text, time_created integer not null, time_updated integer not null)`)
	execSQL(t, db, `create table session (id text primary key, project_id text not null, title text not null, time_created integer not null, time_updated integer not null, time_archived integer)`)
	execSQL(t, db, `create table message (id text primary key, session_id text not null, time_created integer not null, time_updated integer not null, data text not null)`)
}

func openSQLite(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	return db
}

func execSQL(t *testing.T, db *sql.DB, statement string) {
	t.Helper()
	if _, err := db.Exec(statement); err != nil {
		t.Fatalf("failed to execute SQL %q: %v", statement, err)
	}
}
