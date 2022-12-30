package handler

import (
	"encoding/json"
	"fmt"
	"github.com/itksb/go-url-shortener/api"
	"github.com/itksb/go-url-shortener/internal/user"
	"io"
	"net/http"
)

// APIShortenURL -.
func (h *Handler) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		SendJSONError(w, "error of reading request", http.StatusInternalServerError)
		h.logger.Error("api shorten request", err)
		return
	}
	request := api.ShortenRequest{}

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

	response := api.ShortenResponse{Result: createShortenURL(sURLId, h.cfg.ShortBaseURL)}
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

// APIShortenURLBatch - .
func (h *Handler) APIShortenURLBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	requestItems := api.ShortenBatchRequest{}
	err := json.NewDecoder(r.Body).Decode(&requestItems)
	if err != nil {
		h.logger.Error(err)
		SendJSONError(w, "bad input json", http.StatusBadRequest)
		return
	}
	if len(requestItems) == 0 {
		SendJSONError(w, "bad input request: empty input", http.StatusBadRequest)
		h.logger.Error("bad request. URL is empty", err)
		return
	}

	ctx := r.Context()
	userID, ok := ctx.Value(user.FieldID).(string)
	if !ok {
		h.logger.Error("no user id found")
		SendJSONError(w, "no user found", http.StatusInternalServerError)
		return
	}

	response := api.ShortenBatchResponse{}
	for _, shortenBatchItemRequest := range requestItems {
		sURLId, err := h.urlshortener.ShortenURL(r.Context(), shortenBatchItemRequest.OriginalURL, userID)
		if err != nil {
			h.logger.Error("ApiShortenUrl. urlshortener.ShortenURL(...) call error", err.Error())
			SendJSONError(w, "shortener service error", http.StatusInternalServerError)
			return
		}
		shortURL := createShortenURL(sURLId, h.cfg.ShortBaseURL)
		responseItem := api.ShortenBatchItemResponse{
			CorrelationID: shortenBatchItemRequest.CorrelationID,
			ShortURL:      shortURL,
		}
		response = append(response, responseItem)
	}

	SendJSONOk(w, response, http.StatusCreated)

}
