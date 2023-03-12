package router

import (
	"compress/gzip"
	"context"
	"encoding/gob"
	"github.com/itksb/go-url-shortener/internal/handler"
	"github.com/itksb/go-url-shortener/internal/user"
	"github.com/itksb/go-url-shortener/pkg/logger"
	its "github.com/itksb/go-url-shortener/pkg/session"
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

const MsgSessionRestoringError = "session restoring error"
const MsgSaveSessionError = "save session error"

// NewAuthMiddleware - setup user context
// Additionally generates UserId and saves it to the cookie and context
// see examples: https://bash-shell.net/blog/dependency-injection-golang-http-middleware/
func NewAuthMiddleware(sessionStore its.Store, l *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gob.Register(user.FieldID) // suddenly ага :) по идее поместить там где тип, но там нету инициализации пока

			userSession, err := sessionStore.Get(r, "s")
			if err != nil {
				handler.SendJSONError(w, MsgSessionRestoringError, http.StatusInternalServerError)
				l.Error(err)
				return
			}

			userID, savedInSession := userSession.Values[user.FieldID].(string)
			if savedInSession { // just set value in context
				*r = *r.WithContext(context.WithValue(r.Context(), user.FieldID, userID))
			} else { // no user in session, then create one
				userID = user.GenerateUserID()
				*r = *r.WithContext(context.WithValue(r.Context(), user.FieldID, userID))
				userSession.Values[user.FieldID] = userID
				err = userSession.Save(r, w)
				if err != nil {
					http.Error(w, MsgSaveSessionError, http.StatusInternalServerError)
					handler.SendJSONError(w, MsgSaveSessionError, http.StatusInternalServerError)
					l.Error(err)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
