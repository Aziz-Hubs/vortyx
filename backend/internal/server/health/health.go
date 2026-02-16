// =============================================================================
// Package: health
// File: health.go
// Purpose: Health check endpoints
// Created: 2026-02-15
// =============================================================================

package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// NewHandler creates a new health check handler.
func NewHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		services := make(map[string]string)

		// Check DB
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			status = "degraded"
			services["database"] = "unreachable"
		} else {
			services["database"] = "ok"
		}

		// We could add more checks (Redis, MQ, etc.)

		resp := HealthStatus{
			Status:    status,
			Timestamp: time.Now(),
			Services:  services,
		}

		w.Header().Set("Content-Type", "application/json")
		if status != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(resp)
	}
}
