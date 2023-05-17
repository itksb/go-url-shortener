package handler

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_createShortenURL(t *testing.T) {
	id := "abcd1234"
	baseURL := "https://example.com"
	expectedResult := "https://example.com/abcd1234"
	result := createShortenURL(id, baseURL)

	if result != expectedResult {
		t.Errorf("Expected result: %s, but got: %s", expectedResult, result)
	}
}

func TestSendJSONError(t *testing.T) {
	rr := httptest.NewRecorder()
	code := http.StatusBadRequest
	msg := "Bad Request"
	SendJSONError(rr, msg, code)

	if rr.Code != code {
		t.Errorf("expected response code %d, but got %d", code, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type 'application/json', but got '%s'", contentType)
	}

	type exppectedBodyContent struct {
		Error string `json:"error"`
	}

	expectedBody := exppectedBodyContent{
		Error: msg,
	}
	gotBody := exppectedBodyContent{}
	err := json.Unmarshal(rr.Body.Bytes(), &gotBody)
	if err != nil {
		t.Error("error unmarshalling content")
	}
	assert.Equal(t, expectedBody, gotBody)
}

func TestSendJSONOk(t *testing.T) {
	rr := httptest.NewRecorder()
	content := map[string]string{"ku-ku": "mir"}
	code := http.StatusOK
	err := SendJSONOk(rr, content, code)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rr.Code != code {
		t.Errorf("expected response code %d, but got %d", code, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type 'application/json', but got '%s'", contentType)
	}

	type expectedResponse struct {
		Kuku string `json:"ku-ku"`
	}

	expectedBody := expectedResponse{
		Kuku: "mir",
	}

	actualresp := expectedResponse{}
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &actualresp))
	assert.Equal(t, expectedBody, actualresp)

}

func TestSendJSONOk2(t *testing.T) {
	rr := httptest.NewRecorder()
	content := "some content here"
	code := http.StatusOK
	err := SendJSONOk(rr, content, code)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rr.Code != code {
		t.Errorf("expected response code %d, but got %d", code, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type 'application/json', but got '%s'", contentType)
	}

	expectedBody := DefaultResponse{Msg: content}

	actualresp := DefaultResponse{}
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &actualresp))
	assert.Equal(t, expectedBody, actualresp)

}
