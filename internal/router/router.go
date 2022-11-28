package router

import (
	"github.com/gorilla/mux"
	"github.com/itksb/go-url-shortener/internal/handler"
	"net/http"
)

// NewRouter - constructor
func NewRouter(h *handler.Handler) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", h.ShortenURL).Methods(http.MethodPost)
	r.HandleFunc("/{id:[0-9]+}", h.GetURL).Methods(http.MethodGet)
	// api
	r.HandleFunc("/api/shorten", h.APIShortenURL).Methods(http.MethodPost)

	r.HandleFunc("/health", h.HealthCheck).Methods(http.MethodGet)

	r.Use(gzipMiddleware)

	return r
}
