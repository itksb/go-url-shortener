package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

// APIShortenURL -.
func (h *Handler) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("api shorten request", err)
		return
	}
	var request shortenRequest = shortenRequest{}

	if err := json.Unmarshal(reqBytes, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Error("ApiShortenUrl. Bad request. Json Unmarshalling error", err)
		return
	}

	if request.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Error("ApiShortenUrl. Bad request. URL is empty", err)
		return
	}

	sURLId, err := h.urlshortener.ShortenURL(r.Context(), request.URL)
	if err != nil {
		h.logger.Error("Shorten url failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := shortenResponse{Result: createShortenURL(sURLId, h.cfg.ShortBaseURL)}
	if err := encoder.Encode(response); err != nil {
		h.logger.Error("Encoding to json error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
