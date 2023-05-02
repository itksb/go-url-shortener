package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/itksb/go-url-shortener/api"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/dbstorage"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/internal/user"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

//goland:noinspection HttpUrlsUsage
func TestHandler_ApiShortenURL(t *testing.T) {
	type fields struct {
		logger       logger.Interface
		urlshortener *shortener.Service
		cfg          config.Config
	}

	type args struct {
		method string
		target string
		body   io.Reader
		userID string
	}

	type want struct {
		code        int
		responseURL string
		contentType string
	}

	l, err := logger.NewLogger()
	if err != nil {
		t.Error("Could not create Logger instance!")
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Positive test",
			fields: fields{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]shortener.URLListItem{
					1: {
						OriginalURL: "http://shorten.ru",
						UserID:      "7c7bf38e-a76f-4640-acac-c0bb680b68e4",
					},
				})),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body:   strings.NewReader(`{"url":"http://some.url"}`),
				userID: "1",
			},
			want: want{
				code:        http.StatusCreated,
				responseURL: "http://short.base/0",
				contentType: "application/json",
			},
		},
		{
			name: "Negative test",
			fields: fields{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(
					map[int64]shortener.URLListItem{1: {
						OriginalURL: "http://shorten.ru",
						UserID:      "1",
					}},
				)),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body:   strings.NewReader(`{"url":""}`),
				userID: "1",
			},
			want: want{
				code:        http.StatusBadRequest,
				responseURL: "http://short.base/0",
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				logger:       tt.fields.logger,
				urlshortener: tt.fields.urlshortener,
				cfg:          tt.fields.cfg,
			}

			writer := httptest.NewRecorder()
			request := httptest.NewRequest(
				tt.args.method,
				tt.args.target,
				tt.args.body,
			)

			*request = *request.WithContext(context.WithValue(request.Context(), user.FieldID, tt.args.userID))

			h.APIShortenURL(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if tt.want.code == http.StatusBadRequest && res.StatusCode == http.StatusBadRequest {
				if len(body) != 0 {
					var apiError APIError
					err = json.Unmarshal(body, &apiError)
					if err != nil || apiError.Error == "" {
						t.Fatalf("Status is %d (bad request), but body does not contain error", http.StatusBadRequest)
					}
				}
				return
			}

			if len(body) == 0 {
				t.Fatalf("Empty response: response body is empty!")
			}

			apiResp := struct {
				URL string `json:"result"`
			}{}

			if err := json.Unmarshal(body, &apiResp); err != nil {
				t.Fatalf("Expected answer should be correct json. Request body: %s. Got body: %s. Error: %s", tt.args.body, body, err.Error())
			}

			if apiResp.URL != tt.want.responseURL {
				t.Errorf("Expected url=%s, got %s", tt.want.responseURL, writer.Body.String())
			}

			// заголовок ответа
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}

		})
	}
}

func TestHandler_APIListUserURL(t *testing.T) {

	t.Run("positive", func(t *testing.T) {

		// create a mock context with user ID
		userID := "1"
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		createdTime := time.Date(1986, time.August, 25, 0, 0, 0, 0, time.Local)

		storage := newStorageMock(map[int64]shortener.URLListItem{
			1: {
				ID:          1,
				UserID:      userID,
				ShortURL:    "http://short.example.com/1",
				OriginalURL: "http://example.com",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
			2: {
				ID:          2,
				UserID:      userID,
				ShortURL:    "http://short.example.com/2",
				OriginalURL: "http://example.com/qwerty",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
		})
		urlshortener := shortener.NewShortener(l, storage)

		h := NewHandler(
			l,
			urlshortener,
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{
				AppPort:      80,
				AppHost:      "http://localhost.com",
				ShortBaseURL: "http://short.example.com",
			},
		)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/user/url", nil)
		req = req.WithContext(ctx)
		require.NoError(t, err)

		// call the APIListUserURL method
		h.APIListUserURL(rr, req)
		// check that the response status code is OK
		assert.Equal(t, http.StatusOK, rr.Code)

		// check that the response body contains the expected data
		expectedResp := []shortener.URLListItem{
			{
				ID:          1,
				UserID:      userID,
				ShortURL:    "http://short.example.com/1",
				OriginalURL: "http://example.com",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
			},
			{
				ID:          2,
				UserID:      userID,
				ShortURL:    "http://short.example.com/2",
				OriginalURL: "http://example.com/qwerty",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
			},
		}
		expectedJSON, err := json.Marshal(expectedResp)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), rr.Body.String())

	})

	t.Run("positive one item", func(t *testing.T) {

		// create a mock context with user ID
		userID := "user_id"
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		createdTime := time.Date(1986, time.August, 25, 0, 0, 0, 0, time.Local)

		storage := newStorageMock(map[int64]shortener.URLListItem{
			1: {
				ID:          1,
				UserID:      userID,
				ShortURL:    "http://short.example.com/1",
				OriginalURL: "http://example.com",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
			2: {
				ID:          2,
				UserID:      "userID",
				ShortURL:    "http://short.example.com/2",
				OriginalURL: "http://example.com/qwerty",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
		})
		urlshortener := shortener.NewShortener(l, storage)

		h := NewHandler(
			l,
			urlshortener,
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{
				AppPort:      80,
				AppHost:      "http://localhost.com",
				ShortBaseURL: "http://short.example.com",
			},
		)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/user/url", nil)
		req = req.WithContext(ctx)
		require.NoError(t, err)

		// call the APIListUserURL method
		h.APIListUserURL(rr, req)
		// check that the response status code is OK
		assert.Equal(t, http.StatusOK, rr.Code)

		// check that the response body contains the expected data
		expectedResp := []shortener.URLListItem{
			{
				ID:          1,
				UserID:      userID,
				ShortURL:    "http://short.example.com/1",
				OriginalURL: "http://example.com",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
			},
		}
		expectedJSON, err := json.Marshal(expectedResp)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), rr.Body.String())

	})

	t.Run("wrong request type - post", func(t *testing.T) {

		// create a mock context with user ID
		userID := "user_id"
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		storage := newStorageMock(map[int64]shortener.URLListItem{})
		urlshortener := shortener.NewShortener(l, storage)

		h := NewHandler(
			l,
			urlshortener,
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{},
		)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/user/url", nil)
		req = req.WithContext(ctx)
		require.NoError(t, err)

		// call the APIListUserURL method
		h.APIListUserURL(rr, req)
		// check that the response status code is OK
		assert.Equal(t, http.StatusNoContent, rr.Code)

	})
}

func TestHandler_APIShortenURLBatch(t *testing.T) {

	t.Run("positive", func(tt *testing.T) {

		ctrl := gomock.NewController(tt)
		defer func() {
			if ctrl != nil {
				ctrl.Finish()
			}
		}()

		userID := "1"

		storage := newStorageMock(map[int64]shortener.URLListItem{})

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		urlshortener := shortener.NewShortener(l, storage)

		h := NewHandler(
			l,
			urlshortener,
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{
				AppPort:      80,
				AppHost:      "http://localhost.com",
				ShortBaseURL: "https://short.test",
			},
		)

		// Prepare the request body
		body, _ := json.Marshal(api.ShortenBatchRequest{
			{
				CorrelationID: "3",
				OriginalURL:   "https://google.com",
			},
			{
				CorrelationID: "4",
				OriginalURL:   "https://facebook.com",
			},
		})

		// Set up the request and response
		req, err := http.NewRequest("POST", "/shorten/batch", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req = req.WithContext(context.WithValue(req.Context(), user.FieldID, userID))
		rr := httptest.NewRecorder()

		// Call the handler's APIShortenURLBatch method
		h.APIShortenURLBatch(rr, req)

		// Verify that the response status code is 201
		assert.Equal(tt, http.StatusCreated, rr.Code)

		// Verify that the response body contains the correct short URLs
		expectedBody, _ := json.Marshal(api.ShortenBatchResponse{
			{
				CorrelationID: "3",
				ShortURL:      "https://short.test/0",
			},
			{
				CorrelationID: "4",
				ShortURL:      "https://short.test/1",
			},
		})
		assert.JSONEq(tt, string(expectedBody), rr.Body.String())

	})

	t.Run("no user found", func(tt *testing.T) {

		ctrl := gomock.NewController(tt)
		defer func() {
			if ctrl != nil {
				ctrl.Finish()
			}
		}()

		storage := newStorageMock(map[int64]shortener.URLListItem{})

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		urlshortener := shortener.NewShortener(l, storage)

		h := NewHandler(
			l,
			urlshortener,
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{
				ShortBaseURL: "https://short.test",
			},
		)

		// Prepare the request body
		body, _ := json.Marshal(api.ShortenBatchRequest{
			{
				CorrelationID: "3",
				OriginalURL:   "https://google.com",
			},
			{
				CorrelationID: "4",
				OriginalURL:   "https://facebook.com",
			},
		})

		// Set up the request and response
		req, err := http.NewRequest("POST", "/shorten/batch", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		// Call the handler's APIShortenURLBatch method
		h.APIShortenURLBatch(rr, req)

		// Verify that the response status code is 201
		assert.Equal(tt, http.StatusInternalServerError, rr.Code)

		// Verify that the response body contains the correct short URLs
		expectedBody, _ := json.Marshal(APIError{Error: "no user found"})
		assert.JSONEq(tt, string(expectedBody), rr.Body.String())

	})

	t.Run("empty request error", func(tt *testing.T) {

		ctrl := gomock.NewController(tt)
		defer func() {
			if ctrl != nil {
				ctrl.Finish()
			}
		}()

		userID := "1"

		storage := newStorageMock(map[int64]shortener.URLListItem{})

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		urlshortener := shortener.NewShortener(l, storage)

		h := NewHandler(
			l,
			urlshortener,
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{
				AppPort:      80,
				AppHost:      "http://localhost.com",
				ShortBaseURL: "https://short.test",
			},
		)

		// Prepare the request body
		body, _ := json.Marshal(api.ShortenBatchRequest{})

		// Set up the request and response
		req, err := http.NewRequest("POST", "/shorten/batch", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req = req.WithContext(context.WithValue(req.Context(), user.FieldID, userID))
		rr := httptest.NewRecorder()

		// Call the handler's APIShortenURLBatch method
		h.APIShortenURLBatch(rr, req)

		// Verify that the response status code is 201
		assert.Equal(tt, http.StatusBadRequest, rr.Code)

		// Verify that the response body contains the correct short URLs
		expectedBody, _ := json.Marshal(APIError{Error: "bad input request: empty input"})
		assert.JSONEq(tt, string(expectedBody), rr.Body.String())

	})

}

func TestHandler_APIDeleteURLBatch(t *testing.T) {

	t.Run("success", func(tt *testing.T) {
		ctrl := gomock.NewController(tt)
		defer func() {
			if ctrl != nil {
				ctrl.Finish()
			}
		}()

		userID := "1"
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		createdTime := time.Date(1986, time.August, 25, 0, 0, 0, 0, time.Local)

		storage := newStorageMock(map[int64]shortener.URLListItem{
			1: {
				ID:          1,
				UserID:      userID,
				ShortURL:    "http://short.example.com/1",
				OriginalURL: "http://example.com",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
			2: {
				ID:          2,
				UserID:      userID,
				ShortURL:    "http://short.example.com/2",
				OriginalURL: "http://example.com/qwerty",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
		})

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		h := NewHandler(
			l,
			shortener.NewShortener(l, storage),
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{},
		)

		// создаем входные данные
		ids := []string{"1", "2"}
		jsonData, err := json.Marshal(ids)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/user/urls", bytes.NewBuffer(jsonData))
		require.NoError(t, err)

		// вызываем тестируемый метод
		rr := httptest.NewRecorder()
		h.APIDeleteURLBatch(rr, req)

		// проверяем результат
		assert.Equal(t, http.StatusAccepted, rr.Code)
	})

	t.Run("no user", func(tt *testing.T) {
		ctrl := gomock.NewController(tt)
		defer func() {
			if ctrl != nil {
				ctrl.Finish()
			}
		}()

		ctx := context.Background()
		storage := newStorageMock(map[int64]shortener.URLListItem{})

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		h := NewHandler(
			l,
			shortener.NewShortener(l, storage),
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{},
		)

		// создаем входные данные
		ids := []string{"1", "2"}
		jsonData, err := json.Marshal(ids)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/user/urls", bytes.NewBuffer(jsonData))
		require.NoError(t, err)

		// вызываем тестируемый метод
		rr := httptest.NewRecorder()
		h.APIDeleteURLBatch(rr, req)

		// проверяем результат
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("no ids exist", func(tt *testing.T) {
		ctrl := gomock.NewController(tt)
		defer func() {
			if ctrl != nil {
				ctrl.Finish()
			}
		}()

		userID := "1"
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		createdTime := time.Date(1986, time.August, 25, 0, 0, 0, 0, time.Local)

		storage := newStorageMock(map[int64]shortener.URLListItem{
			1: {
				ID:          1,
				UserID:      userID,
				ShortURL:    "http://short.example.com/1",
				OriginalURL: "http://example.com",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
			2: {
				ID:          2,
				UserID:      userID,
				ShortURL:    "http://short.example.com/2",
				OriginalURL: "http://example.com/qwerty",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
		})

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		h := NewHandler(
			l,
			shortener.NewShortener(l, storage),
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{},
		)

		// создаем входные данные
		ids := []string{"3", "4"}
		jsonData, err := json.Marshal(ids)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/user/urls", bytes.NewBuffer(jsonData))
		require.NoError(t, err)

		// вызываем тестируемый метод
		rr := httptest.NewRecorder()
		h.APIDeleteURLBatch(rr, req)

		// проверяем результат
		assert.Equal(t, http.StatusAccepted, rr.Code)
	})

	t.Run("id collection is empty", func(tt *testing.T) {
		ctrl := gomock.NewController(tt)
		defer func() {
			if ctrl != nil {
				ctrl.Finish()
			}
		}()

		userID := "1"
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		createdTime := time.Date(1986, time.August, 25, 0, 0, 0, 0, time.Local)

		storage := newStorageMock(map[int64]shortener.URLListItem{
			1: {
				ID:          1,
				UserID:      userID,
				ShortURL:    "http://short.example.com/1",
				OriginalURL: "http://example.com",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
			2: {
				ID:          2,
				UserID:      userID,
				ShortURL:    "http://short.example.com/2",
				OriginalURL: "http://example.com/qwerty",
				CreatedAt:   createdTime.String(),
				UpdatedAt:   createdTime.String(),
				DeletedAt:   nil,
			},
		})

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		h := NewHandler(
			l,
			shortener.NewShortener(l, storage),
			&dbstorage.Storage{},
			&dbstorage.Storage{},
			config.Config{},
		)

		// создаем входные данные
		ids := []string{}
		jsonData, err := json.Marshal(ids)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/user/urls", bytes.NewBuffer(jsonData))
		require.NoError(t, err)

		// вызываем тестируемый метод
		rr := httptest.NewRecorder()
		h.APIDeleteURLBatch(rr, req)

		// проверяем результат
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

}
