package handler

import (
	"github.com/itksb/go-url-shortener/internal/user"
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

	ctx := r.Context()
	userID, ok := ctx.Value(user.FieldID).(string)
	if !ok {
		h.logger.Error("no user id found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sURLId, err := h.urlshortener.ShortenURL(r.Context(), inURL, userID)
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

	if len(originalURL) == 0 {
		h.logger.Info("Url not found for id:", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

}
