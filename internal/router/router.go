package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/itksb/go-url-shortener/internal/handler"
	"github.com/itksb/go-url-shortener/pkg/session"
	"net/http"
)

// NewRouter - constructor
func NewRouter(h *handler.Handler, sessionStore session.Store) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(gzipUnpackMiddleware)
	authMdl := NewAuthMiddleware(sessionStore)
	r.Use(authMdl)
	r.Use(gzipMiddleware)

	r.MethodFunc(http.MethodPost, "/", h.ShortenURL)
	r.MethodFunc(http.MethodGet, "/{id:[0-9]+}", h.GetURL)
	// api
	r.MethodFunc(http.MethodPost, "/api/shorten", h.APIShortenURL)
	r.MethodFunc(http.MethodGet, "/health", h.HealthCheck)

	return r, nil
}
