package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/itksb/go-url-shortener/internal/handler"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/itksb/go-url-shortener/pkg/session"
	"net/http"
)

// NewRouter - constructor
func NewRouter(h *handler.Handler, sessionStore session.Store, l *logger.Logger, debug bool) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(gzipUnpackMiddleware)
	authMdl := NewAuthMiddleware(sessionStore, l)
	r.Use(authMdl)
	r.Use(gzipMiddleware)

	r.MethodFunc(http.MethodPost, "/", h.ShortenURL)
	r.MethodFunc(http.MethodGet, "/{id:[0-9]+}", h.GetURL)

	r.Group(func(r2 chi.Router) {
		// apply CORS middleware for api routes
		r2.Use(NewCors())
		// api routes
		r2.MethodFunc(http.MethodPost, "/api/shorten", h.APIShortenURL)
		r2.MethodFunc(http.MethodGet, "/api/user/urls", h.APIListUserURL)
		r2.MethodFunc(http.MethodPost, "/api/shorten/batch", h.APIShortenURLBatch)
		r2.MethodFunc(http.MethodDelete, "/api/user/urls", h.APIDeleteURLBatch)
	})

	r.MethodFunc(http.MethodGet, "/health", h.HealthCheck)
	r.MethodFunc(http.MethodGet, "/ping", h.Ping)

	if debug {
		r.Mount("/debug", middleware.Profiler())
		l.Info("enables profiler route (due to debug environment): /debug")
	}

	return r, nil
}
