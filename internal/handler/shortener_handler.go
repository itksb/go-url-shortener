package handler

import (
	"io"
	"net/http"
	"strings"
)

// ShortenURL -.
func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("shorten request", err)
		return
	}
	inURL := string(bytes)

	sURLId, err := h.urlshortener.ShortenURL(r.Context(), inURL)
	if err != nil {
		h.logger.Error("Shorten url failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")

	w.Write([]byte(createShortenURL(sURLId, h.cfg.ShortBaseURL)))

}

// GetURL - endpoint handler
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	_, id, ok := strings.Cut(r.URL.Path, "/")
	if !ok {
		h.logger.Error("parse url id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, err := h.urlshortener.GetURL(r.Context(), id)
	if err != nil {
		h.logger.Info("Id not found", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

}
