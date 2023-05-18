package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/itksb/go-url-shortener/api"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/dbstorage"
	"github.com/itksb/go-url-shortener/internal/shortener"
	mock_shortener "github.com/itksb/go-url-shortener/internal/shortener/mock"
	"github.com/itksb/go-url-shortener/internal/user"
	"github.com/itksb/go-url-shortener/pkg/logger"
	mock_logger "github.com/itksb/go-url-shortener/pkg/logger/mock"
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
func TestHandler_APIShortenURL(t *testing.T) {
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
					0: {
						OriginalURL: "http://shorten.ru",
						UserID:      "1",
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
				responseURL: "http://short.base/1",
				contentType: "application/json",
			},
		},

		{
			name: "Negative: empty url",
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

		{
			name: "negative: empty user id",
			fields: fields{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]shortener.URLListItem{
					0: {
						OriginalURL: "http://shorten.ru",
						UserID:      "1",
					},
				})),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body:   strings.NewReader(`{"url":"http://some.url"}`),
				userID: "",
			},
			want: want{
				code:        http.StatusInternalServerError,
				responseURL: "http://short.base/1",
				contentType: "application/json",
			},
		},

		{
			name: "Empty request body",
			fields: fields{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]shortener.URLListItem{
					0: {
						OriginalURL: "http://shorten.ru",
						UserID:      "1",
					},
				})),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body: func() *strings.Reader {
					// create a new io.Reader and close it to simulate a closed request body
					r := strings.NewReader(`{"url":"http://some.url"}`)
					r.Reset("")
					return r
				}(),
				userID: "1",
			},
			want: want{
				code:        http.StatusInternalServerError,
				responseURL: "http://short.base/1",
				contentType: "application/json",
			},
		},

		{
			name: "Bad input request body",
			fields: fields{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]shortener.URLListItem{
					0: {
						OriginalURL: "http://shorten.ru",
						UserID:      "1",
					},
				})),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body:   strings.NewReader(`{"url":"http://some.url"`),
				userID: "1",
			},
			want: want{
				code:        http.StatusBadRequest,
				responseURL: "http://short.base/1",
				contentType: "application/json",
			},
		},

		{
			name: "shortener service error wrong input struct",
			fields: fields{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]shortener.URLListItem{
					0: {
						OriginalURL: "http://shorten.ru",
						UserID:      "1",
					},
				})),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body:   strings.NewReader(`{"url1":"http://some.url"}`),
				userID: "1",
			},
			want: want{
				code:        http.StatusBadRequest,
				responseURL: "http://short.base/1",
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

			if tt.want.code == http.StatusInternalServerError {
				return
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

		ctrl := gomock.NewController(t)
		storage := mock_shortener.NewMockShortenerStorage(ctrl)
		storage.EXPECT().ListURLByUserID(ctx, userID).Return([]shortener.URLListItem{
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
		}, nil)

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

	t.Run("no user found in the session", func(t *testing.T) {

		// create a mock context with user ID
		userID := ""
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		ctrl := gomock.NewController(t)
		l := mock_logger.NewMockInterface(ctrl)

		l.EXPECT().Error("user id not found, but it must already be here. see middleware which setup user session")

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
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		// check that the response body contains the expected data
		expectedResp := map[string]interface{}(map[string]interface{}{"error": "no user found in the session"})
		expectedJSON, err := json.Marshal(expectedResp)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), rr.Body.String())

	})

	t.Run("error while searching user urls", func(t *testing.T) {

		// create a mock context with user ID
		userID := "1"
		ctx := context.WithValue(context.Background(), user.FieldID, userID)

		l, err := logger.NewLogger()
		if err != nil {
			t.Errorf("error calling logger.NewLogger, %s", err.Error())
		}

		ctrl := gomock.NewController(t)
		storage := mock_shortener.NewMockShortenerStorage(ctrl)

		storage.EXPECT().ListURLByUserID(ctx, userID).Return(
			[]shortener.URLListItem{},
			errors.New("error while searching user urls"))

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
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

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

	t.Run("bad input json", func(tt *testing.T) {

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
		body := []byte(" [ { bad json :")

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
		expectedBody, _ := json.Marshal(
			map[string]interface{}(map[string]interface{}{"error": "bad input json"}),
		)
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

	t.Run("bad input json", func(tt *testing.T) {
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
		jsonData := []byte(`["foo":"bar"}`)
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

func TestHandler_APIInternalStats(t *testing.T) {
	l, err := logger.NewLogger()
	assert.NoError(t, err)

	createdTime := time.Date(1986, time.August, 25, 0, 0, 0, 0, time.Local)
	storage := newStorageMock(map[int64]shortener.URLListItem{
		1: {
			ID:          1,
			UserID:      "1",
			ShortURL:    "http://short.example.com/1",
			OriginalURL: "http://example.com",
			CreatedAt:   createdTime.String(),
			UpdatedAt:   createdTime.String(),
		},
		2: {
			ID:          2,
			UserID:      "1",
			ShortURL:    "http://short.example.com/2",
			OriginalURL: "http://example.com/qwerty",
			CreatedAt:   createdTime.String(),
			UpdatedAt:   createdTime.String(),
		},
		3: {
			ID:          3,
			UserID:      "1",
			ShortURL:    "http://short.example.com/3",
			OriginalURL: "http://example.com/qwerty",
			CreatedAt:   createdTime.String(),
			UpdatedAt:   createdTime.String(),
		},
	})

	urlshortener := shortener.NewShortener(l, storage)
	// Создаем объект Handler
	handler := NewHandler(l, urlshortener, &dbstorage.Storage{}, nil, config.Config{})

	// Создаем http.Request
	req, err := http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// Создаем ResponseWriter с поддержкой записи тела ответа
	w := httptest.NewRecorder()

	// Вызываем функцию APIInternalStats с созданными объектами
	handler.APIInternalStats(w, req)

	// Проверяем код ответа
	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}

	// Проверяем содержимое ответа
	var response api.ShortenInternalStatsResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	expected := api.ShortenInternalStatsResponse{Urls: 3, Users: 1}
	if response != expected {
		t.Errorf("expected response %v; got %v", expected, response)
	}
}

func TestHandler_APIShortenURL_EmptyBody(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer func() {
		if ctrl != nil {
			ctrl.Finish()
		}
	}()

	l := mock_logger.NewMockInterface(ctrl)

	createdTime := time.Date(1986, time.August, 25, 0, 0, 0, 0, time.Local)
	storage := newStorageMock(map[int64]shortener.URLListItem{
		10: {
			ID:          1,
			UserID:      "1",
			ShortURL:    "http://short.example.com/1",
			OriginalURL: "http://example.com",
			CreatedAt:   createdTime.String(),
			UpdatedAt:   createdTime.String(),
		},
		20: {
			ID:          2,
			UserID:      "1",
			ShortURL:    "http://short.example.com/2",
			OriginalURL: "http://example.com/qwerty",
			CreatedAt:   createdTime.String(),
			UpdatedAt:   createdTime.String(),
		},
	})

	urlshortener := shortener.NewShortener(l, storage)

	h := &Handler{
		logger:       l,
		urlshortener: urlshortener,
		cfg:          config.Config{},
	}

	writer := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/shorten",
		strings.NewReader(``),
		//strings.NewReader(`{"url":"http://some.url"}`),
	)

	*request = *request.WithContext(context.WithValue(request.Context(), user.FieldID, "1"))

	l.EXPECT().Error("api shorten request: empty body")

	h.APIShortenURL(writer, request)
	res := writer.Result()
	defer res.Body.Close()

}
