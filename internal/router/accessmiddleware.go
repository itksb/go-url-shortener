package router

import (
	"fmt"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"net"
	"net/http"
	"strings"
)

const accessForbiddenText = "Internal resource. Access forbidden by security policy"

func NewAccessMiddleware(trustedSubnet string, l *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				http.Error(w, accessForbiddenText, http.StatusForbidden)
				return
			}

			userIP := r.Header.Get("X-Real-IP")
			if userIP == "" {
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				fmt.Fprintln(w, accessForbiddenText)

				return
			}

			parsedIP := net.ParseIP(userIP)

			if !strings.Contains(trustedSubnet, "/") {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintln(w, "server misconfiguration error")

				return
			}

			_, configSubnet, err := net.ParseCIDR(trustedSubnet)

			if err != nil {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintln(w, "server misconfiguration error")

				return
			}

			if !configSubnet.Contains(parsedIP) {

				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintln(w, accessForbiddenText)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
