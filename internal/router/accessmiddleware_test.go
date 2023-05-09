package router

import (
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewAccessMiddleware - tests
func TestNewAccessMiddleware(t *testing.T) {

	l, err := logger.NewLogger()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name          string
		headers       map[string]string
		expected      int
		trustedSubnet string
	}{
		{
			name: "valid request",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.10",
			},
			expected:      http.StatusOK,
			trustedSubnet: "192.168.1.0/24",
		},
		{
			name: "empty subnets",
			headers: map[string]string{
				"X-Real-IP": "",
			},
			expected:      http.StatusForbidden,
			trustedSubnet: "192.168.1.0/24",
		},
		{
			name: "invalid subnet",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.1",
			},
			expected:      http.StatusForbidden,
			trustedSubnet: "192.168.1.0",
		},
		{
			name: "empty trusted subnet",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.10",
			},
			expected:      http.StatusForbidden,
			trustedSubnet: "",
		},
	}

	// Запускаем тесты для middleware
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			accessMiddleware := NewAccessMiddleware(tt.trustedSubnet, l)
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			handler := accessMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			for key, value := range tt.headers {
				request.Header.Set(key, value)
			}

			handler.ServeHTTP(w, request)
			assert.Equal(t, tt.expected, w.Code)

		})
	}
}
