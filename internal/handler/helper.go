package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateShortenURL - creates shorten url
func CreateShortenURL(id string, baseURL string) string {
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

// CreateErrorMsg - creates APIError with msg
func CreateErrorMsg(msg string) ([]byte, error) {
	e := APIError{Error: msg}
	js, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return nil, err
	}

	return js, nil
}

// SendJSONError - creates APIError with msg, sets json headers
func SendJSONError(w http.ResponseWriter, msg string, code int) {
	js, err := CreateErrorMsg(msg)
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

	switch v := content.(type) {
	case string:
		js, err = json.MarshalIndent(DefaultResponse{Msg: v}, "", "  ")
	default:
		js, err = json.MarshalIndent(content, "", "  ")
	}
	if err != nil {
		http.Error(w, "{\"error\": \"error of json encoding\"}", http.StatusInternalServerError)
	}

	if code != http.StatusNoContent {
		w.Header().Add("Content-Type", "application/json")
	}
	w.WriteHeader(code)

	if code != http.StatusNoContent {
		_, err = w.Write(js)
	}
	if err != nil {
		http.Error(w, "{\"error\": \"error sending response\"}", http.StatusInternalServerError)
	}
	return err
}
