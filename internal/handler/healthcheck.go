package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HealthCheck - for monitoring stuff
func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// Ping - checks whether db service is available or not
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if h.dbservice.Ping(ctx) {
		w.WriteHeader(http.StatusOK)
	} else {
		h.logger.Error("db service ping error")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
