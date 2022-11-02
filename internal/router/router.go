package router

import (
	"github.com/itksb/go-url-shortener/internal/handler"
	"net/http"
	"strings"
)

// NewRouter - constructor
func NewRouter(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			h.ShortenURL(writer, request)
			return

		case http.MethodGet:
			_, after, found := strings.Cut(request.URL.Path, "/")
			if found && len(after) > 0 {
				h.GetURL(writer, request)
				return
			}
			fallthrough
		default:
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
	})

	mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		h.HealthCheck(writer, request)
	})

	return mux
}
