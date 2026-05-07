package main

import (
	"log"
	"net/http"

	"github.com/ccmetz/agent-dash/backend/internal/api"
)

func main() {
	log.Println("agent dash backend listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", api.NewServer()))
}
