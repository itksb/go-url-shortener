package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func createShortenURL(id string, baseURL string) string {
	return fmt.Sprintf("%s/%s", baseURL, id)
}

// APIError - base struct for error response
type APIError struct {
	Error string `json:"error"`
}

// DefaultResponse - default response with text message
type DefaultResponse struct {
	Msg string `json:"msg"`
}

// SendJSONError - creates APIError with msg, sets json headers
func SendJSONError(w http.ResponseWriter, msg string, code int) {
	e := APIError{Error: msg}
	js, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		http.Error(w, "{\"error\": \"error of json encoding\"}", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, "{\"error\": \"error sending response\"}", http.StatusInternalServerError)
		return
	}
}

// SendJSONOk sends json response with json header. Creates DefaultResponse response if content is string
func SendJSONOk(w http.ResponseWriter, content interface{}, code int) error {
	var js []byte
	var err error

	switch content.(type) {
	case string:
		js, err = json.MarshalIndent(DefaultResponse{Msg: content.(string)}, "", "  ")
	default:
		js, err = json.MarshalIndent(content, "", "  ")
	}
	if err != nil {
		http.Error(w, "{\"error\": \"error of json encoding\"}", http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err = w.Write(js)
	if err != nil {
		http.Error(w, "{\"error\": \"error sending response\"}", http.StatusInternalServerError)
	}
	return err
}
