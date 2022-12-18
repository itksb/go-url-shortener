package handler

import (
	"encoding/json"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/user"
	"io"
	"net/http"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

type userURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// APIShortenURL -.
func (h *Handler) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		SendJSONError(w, "error of reading request", http.StatusInternalServerError)
		h.logger.Error("api shorten request", err)
		return
	}
	request := shortenRequest{}

	if err := json.Unmarshal(reqBytes, &request); err != nil {
		SendJSONError(w, "bad input request", http.StatusBadRequest)
		h.logger.Error("ApiShortenUrl. Bad request. Json Unmarshalling error", err)
		return
	}

	if request.URL == "" {
		SendJSONError(w, "bad input request: URL is empty", http.StatusBadRequest)
		h.logger.Error("ApiShortenUrl. Bad request. URL is empty", err)
		return
	}

	ctx := r.Context()
	userID, ok := ctx.Value(user.FieldID).(string)
	if !ok {
		h.logger.Error("no user id found")
		SendJSONError(w, "no user found", http.StatusInternalServerError)
		return
	}

	sURLId, err := h.urlshortener.ShortenURL(r.Context(), request.URL, userID)
	if err != nil {
		h.logger.Error("ApiShortenUrl. urlshortener.ShortenURL(...) call error", err.Error())
		SendJSONError(w, "shortener service error", http.StatusInternalServerError)
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

// APIListUserURL = .
func (h *Handler) APIListUserURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(user.FieldID).(string)
	if !ok {
		h.logger.Error("user id not found, but it must already be here. see middleware which setup user session")
		SendJSONError(w, "no user found in the session", http.StatusInternalServerError)
		return
	}

	urlListItems, err := h.urlshortener.ListURLByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("error while searching user urls")
		SendJSONError(w, "shortener service error", http.StatusInternalServerError)
		return
	}

	// creating short urls is infrastructure layer responsibility, that`s why it is here
	for idx := range urlListItems {
		urlListItems[idx].ShortURL = createShortenURL(fmt.Sprint(urlListItems[idx].ID), h.cfg.ShortBaseURL)
	}

	if len(urlListItems) > 0 {
		w.WriteHeader(http.StatusOK)
		if err := SendJSONOk(w, urlListItems, http.StatusOK); err != nil {
			h.logger.Error(err)
			http.Error(w, "error creating response", http.StatusInternalServerError)
			return
		}
	} else {
		if err := SendJSONOk(w, urlListItems, http.StatusNoContent); err != nil {
			h.logger.Error(err)
			http.Error(w, "error creating response", http.StatusInternalServerError)
			return
		}
	}
}
