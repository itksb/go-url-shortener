package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// IPingableDB - infrastructure layer interface. It is separated from the dbstorage.Storage interface
type IPingableDB interface {
	Ping(ctx context.Context) bool
}

// HealthCheck - for monitoring stuff
func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	if err != nil {
		h.logger.Error(err)
	}
}

// Ping - checks whether db service is available or not
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if h.dbping.Ping(ctx) {
		w.WriteHeader(http.StatusOK)
	} else {
		h.logger.Error("db service ping error")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
