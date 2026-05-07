package api

import (
	"encoding/json"
	"net/http"

	"github.com/ccmetz/agent-dash/backend/internal/config"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", statusHandler)
	return mux
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.Load()
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		OK                 bool   `json:"ok"`
		AnalyticsStorePath string `json:"analyticsStorePath"`
	}{
		OK:                 true,
		AnalyticsStorePath: config.AnalyticsStorePath,
	})
}
