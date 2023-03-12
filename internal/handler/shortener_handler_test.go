package handler

import (
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

//goland:noinspection HttpUrlsUsage
func TestHandler_GetURL(t *testing.T) {
	type handlerFields struct {
		logger       logger.Interface
		urlshortener *shortener.Service
		cfg          config.Config
	}
	type args struct {
		method string
		target string
		body   io.Reader
	}

	type want struct {
		code           int
		response       string
		contentType    string
		locationHeader string
	}

	// mocks
	l := &loggerMock{}

	// test
	tests := []struct {
		name   string
		fields handlerFields
		args   args
		want   want
	}{
		// Test cases.
		{
			name: "test1",
			fields: struct {
				logger       logger.Interface
				urlshortener *shortener.Service
				cfg          config.Config
			}{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]shortener.URLListItem{1: {
					OriginalURL: "http://shorten.ru",
					UserID:      "1",
				}})),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "GET",
				target: "/1",
				body:   nil,
			},
			want: want{
				code:           307,
				response:       "",
				contentType:    "",
				locationHeader: "http://shorten.ru",
			}},

		{
			name: "test2",
			fields: struct {
				logger       logger.Interface
				urlshortener *shortener.Service
				cfg          config.Config
			}{
				logger: l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]shortener.URLListItem{1: {
					OriginalURL: "http://shorten.ru",
					UserID:      "1",
				}})),
				cfg: config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "GET",
				target: "/2",
				body:   nil,
			},
			want: want{
				code:           400,
				response:       "",
				contentType:    "",
				locationHeader: "",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				logger:       tt.fields.logger,
				urlshortener: tt.fields.urlshortener,
				cfg:          tt.fields.cfg,
			}

			writer := httptest.ResponseRecorder{}
			request := httptest.NewRequest(
				tt.args.method,
				tt.args.target,
				tt.args.body,
			)

			h.GetURL(&writer, request)
			res := writer.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, writer.Body.String())
			}

			// заголовок ответа
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}

			if len(tt.want.locationHeader) > 0 && res.Header.Get("Location") != tt.want.locationHeader {
				t.Errorf("Expected Location %s, got %s", tt.want.contentType, res.Header.Get("Location"))
			}

		})
	}
}

func TestHandler_GetURL2(t *testing.T) {
	// определяем структуру теста
	type want struct {
		logger       logger.Interface
		urlshortener *shortener.Service
		cfg          config.Config
	}

	l := loggerMock{}
	stMock := newStorageMock(map[int64]shortener.URLListItem{

		1: {
			ID:          1,
			UserID:      "1",
			OriginalURL: "http://ya.rutest/1",
		},
		2: {
			ID:          2,
			UserID:      "1",
			OriginalURL: "https://vk.com",
		},
	})

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields want
		args   args
	}{
		{
			name: "positive 1",
			fields: want{
				logger:       &l,
				urlshortener: shortener.NewShortener(&l, stMock),
				cfg: config.Config{
					ShortBaseURL: "http://localhost:8080",
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"/1",
					nil),
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
			h.GetURL(tt.args.w, tt.args.r)
		})
	}
}
