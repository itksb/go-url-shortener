package handler

import (
	"context"
	"encoding/json"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"net/http"
)

// Handler - endpoint handlers
type Handler struct {
	logger       logger.Interface
	urlshortener urlShortener
	cfg          config.Config
}

type urlShortener interface {
	ShortenURL(ctx context.Context, url string) (string, error)
	GetURL(ctx context.Context, id string) (string, error)
}

// NewHandler - constructor
func NewHandler(logger logger.Interface, shortener urlShortener, cfg config.Config) *Handler {
	return &Handler{
		logger:       logger,
		urlshortener: shortener,
		cfg:          cfg,
	}
}

// HealthCheck - for monitoring stuff
func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
