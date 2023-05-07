package handler

import (
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/dbstorage"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_HealthCheck(t *testing.T) {
	type fields struct {
		logger       logger.Interface
		urlshortener *shortener.Service
		cfg          config.Config
		dbservice    *dbstorage.Storage
	}
	type args struct {
		w          http.ResponseWriter
		httpMethod string
		httpBody   string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "healthcheck success",
			fields: fields{
				logger:       nil,
				urlshortener: nil,
				cfg:          config.Config{},
				dbservice:    nil,
			},
			args: args{
				w:          httptest.NewRecorder(),
				httpMethod: "GET",
				httpBody:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				logger:       tt.fields.logger,
				urlshortener: tt.fields.urlshortener,
				cfg:          tt.fields.cfg,
				dbservice:    tt.fields.dbservice,
			}

			r, err := http.NewRequest(tt.args.httpMethod, "/health", strings.NewReader(tt.args.httpBody))
			if err != nil {
				t.Fatal(err)
			}

			h.HealthCheck(tt.args.w, r)

			res := tt.args.w.(*httptest.ResponseRecorder).Result()
			assert.Equal(t, http.StatusOK, res.StatusCode, "Status code should be 200")
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = res.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			var response map[string]bool
			err = json.Unmarshal(body, &response)
			if err != nil {
				t.Fatal(err)
			}
			assert.True(t, response["ok"], "Response should contain 'ok:true'")
		})
	}
}

func TestHandler_Ping(t *testing.T) {

	t.Run("positive", func(tt *testing.T) {
		ctrl := gomock.NewController(tt)
		// Create a new instance of the mock DB service
		pingDB := NewMockIPingableDB(ctrl)
		// Set up the handler with the mock DB service
		handler := &Handler{dbping: pingDB}

		// Set up the request and response
		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			tt.Fatal(err)
		}
		rr := httptest.NewRecorder()

		// Expect a call to Ping with a context and return true
		pingDB.EXPECT().Ping(gomock.Any()).Return(true)
		// Call the handler's Ping method
		handler.Ping(rr, req)

		// Verify that the response status code is OK
		assert.Equal(tt, http.StatusOK, rr.Code)
	})

	t.Run("if ping false", func(tt *testing.T) {
		ctrl := gomock.NewController(tt)
		// Create a new instance of the mock DB service
		pingDB := NewMockIPingableDB(ctrl)
		l, _ := logger.NewLogger()
		// Set up the handler with the mock DB service
		handler := &Handler{dbping: pingDB, logger: l}

		// Set up the request and response
		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			tt.Fatal(err)
		}
		rr := httptest.NewRecorder()

		// Expect a call to Ping with a context and return false
		pingDB.EXPECT().Ping(gomock.Any()).Return(false)
		// Call the handler's Ping method
		handler.Ping(rr, req)
		assert.Equal(tt, http.StatusInternalServerError, rr.Code)
	})

	t.Run("if POST request", func(tt *testing.T) {
		ctrl := gomock.NewController(tt)
		// Create a new instance of the mock DB service
		pingDB := NewMockIPingableDB(ctrl)
		l, _ := logger.NewLogger()
		// Set up the handler with the mock DB service
		handler := &Handler{dbping: pingDB, logger: l}

		// Set up the request and response
		req, err := http.NewRequest("POST", "/ping", nil)
		if err != nil {
			tt.Fatal(err)
		}
		rr := httptest.NewRecorder()

		// Expect a call to Ping with a context and return false
		pingDB.EXPECT().Ping(gomock.Any()).Return(false)
		// Call the handler's Ping method
		handler.Ping(rr, req)
		assert.Equal(tt, http.StatusInternalServerError, rr.Code)
	})

}
