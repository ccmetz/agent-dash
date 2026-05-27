package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ccmetz/agent-dash/backend/internal/config"
	"github.com/ccmetz/agent-dash/backend/internal/usage"
	_ "modernc.org/sqlite"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", statusHandler)
	mux.HandleFunc("POST /api/usage-sync", syncHandler)
	mux.HandleFunc("GET /api/usage-overview", usageOverviewHandler)
	return mux
}

type statusResponse struct {
	OK                 bool                `json:"ok"`
	AnalyticsStorePath string              `json:"analyticsStorePath"`
	UsageSource        usageSourceResponse `json:"usageSource"`
	UsageSync          usageSyncResponse   `json:"usageSync"`
}

type usageSourceResponse struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Available bool   `json:"available"`
	State     string `json:"state"`
}

type usageSyncResponse struct {
	Status      string             `json:"status"`
	LastRun     *usage.SyncResult  `json:"lastRun,omitempty"`
	RecentRuns  []usage.SyncResult `json:"recentRuns"`
	NextPollAt  string             `json:"nextPollAt"`
	PollSeconds int                `json:"pollSeconds"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.Load()
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	runs, _ := usage.RecentSyncRuns(r.Context(), config.AnalyticsStorePath, 5)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(statusResponse{
		OK:                 true,
		AnalyticsStorePath: config.AnalyticsStorePath,
		UsageSource:        openCodeUsageSourceStatus(config.OpenCodeDatabasePath),
		UsageSync:          syncStatus(runs),
	})
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.Load()
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}
	result, syncErr := usage.SyncOpenCode(r.Context(), config.OpenCodeDatabasePath, config.AnalyticsStorePath)
	w.Header().Set("Content-Type", "application/json")
	if syncErr != nil {
		w.WriteHeader(http.StatusBadGateway)
	}
	_ = json.NewEncoder(w).Encode(result)
}

func usageOverviewHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.Load()
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}
	days := 30
	if value := r.URL.Query().Get("days"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid days", http.StatusBadRequest)
			return
		}
		days = parsed
	}
	end := time.Now().UTC()
	if value := r.URL.Query().Get("end"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			http.Error(w, "invalid end", http.StatusBadRequest)
			return
		}
		end = parsed
	}
	overview, err := usage.UsageOverview(r.Context(), config.AnalyticsStorePath, days, end)
	if err != nil {
		http.Error(w, "failed to load usage overview", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(overview)
}

func syncStatus(runs []usage.SyncResult) usageSyncResponse {
	response := usageSyncResponse{Status: "never_synced", RecentRuns: runs, PollSeconds: 60, NextPollAt: time.Now().UTC().Add(60 * time.Second).Format(time.RFC3339)}
	if len(runs) > 0 {
		response.LastRun = &runs[0]
		response.Status = runs[0].Status
	}
	return response
}

func openCodeUsageSourceStatus(path string) usageSourceResponse {
	usageSource := usageSourceResponse{
		Name:      "OpenCode",
		Path:      path,
		Available: true,
		State:     "available",
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		usageSource.Available = false
		usageSource.State = "missing"
		return usageSource
	} else if err != nil {
		usageSource.Available = false
		usageSource.State = "unavailable"
		return usageSource
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		usageSource.Available = false
		usageSource.State = "unavailable"
		return usageSource
	}
	defer db.Close()

	if err := usage.ValidateOpenCodeSchema(context.Background(), db); err != nil {
		var schemaErr usage.UnsupportedOpenCodeSchemaError
		usageSource.Available = false
		if errors.As(err, &schemaErr) {
			usageSource.State = "unsupported_schema"
		} else {
			usageSource.State = "unavailable"
		}
	}

	return usageSource
}
