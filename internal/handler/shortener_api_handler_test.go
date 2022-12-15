package handler

import (
	"encoding/json"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//goland:noinspection HttpUrlsUsage
func TestHandler_ApiShortenURL(t *testing.T) {
	type fields struct {
		logger       logger.Interface
		urlshortener urlShortener
		cfg          config.Config
	}

	type args struct {
		method string
		target string
		body   io.Reader
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
				logger:       l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]string{1: "http://shorten.ru"})),
				cfg:          config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body:   strings.NewReader(`{"url":"http://some.url"}`),
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
				logger:       l,
				urlshortener: shortener.NewShortener(l, newStorageMock(map[int64]string{1: "http://shorten.ru"})),
				cfg:          config.Config{ShortBaseURL: "http://short.base"},
			},
			args: args{
				method: "POST",
				target: "/api/shorten",
				body:   strings.NewReader(`{"url":""}`),
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
					t.Fatalf("Status is %d (bad request), but body not empty", http.StatusBadRequest)
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
