package router

import (
	"compress/gzip"
	"fmt"
	"github.com/itksb/go-url-shortener/pkg/session"
	"io"
	"log"
	"net/http"
	"strings"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)

	})
}

func gzipUnpackMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// переменная reader будет равна r.Body или *gzip.Reader
		var reader io.ReadCloser
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}
		r.Body = reader

		next.ServeHTTP(w, r)
	})
}

// NewAuthMiddleware
// see examples: https://bash-shell.net/blog/dependency-injection-golang-http-middleware/
func NewAuthMiddleware(sessionStore session.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// проверяет наличие куки сессии,
			// если она есть, значит пытается дешифровать данные из нее и получить UserId
			// если НЕ получилось дешифровать  - создает новую сессию, сохраняет ее в куку
			// если получилось дешифровать, значит создавать и сохранять ничего не надо
			// в конце в любом случае сохраняет ID сессии в контекст и передает дальше

			userSession, err := sessionStore.Get(r, "s")
			if err != nil {
				http.Error(w, "userSession error", http.StatusInternalServerError)
				return
			}

			user, ok := userSession.Values["user"].(string)
			if !ok {
				// user value is not presented here
				//http.Error(w, "userSession error", http.StatusInternalServerError)
				//return
			}

			fmt.Println("user= " + user)

			next.ServeHTTP(w, r)
		})
	}
}
