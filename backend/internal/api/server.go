package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/ccmetz/agent-dash/backend/internal/config"
	"github.com/ccmetz/agent-dash/backend/internal/usage"
	_ "modernc.org/sqlite"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", statusHandler)
	return mux
}

type statusResponse struct {
	OK                 bool                `json:"ok"`
	AnalyticsStorePath string              `json:"analyticsStorePath"`
	UsageSource        usageSourceResponse `json:"usageSource"`
}

type usageSourceResponse struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Available bool   `json:"available"`
	State     string `json:"state"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.Load()
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(statusResponse{
		OK:                 true,
		AnalyticsStorePath: config.AnalyticsStorePath,
		UsageSource:        openCodeUsageSourceStatus(config.OpenCodeDatabasePath),
	})
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
