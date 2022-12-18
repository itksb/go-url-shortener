package app

import (
	"fmt"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/filestorage"
	"github.com/itksb/go-url-shortener/internal/handler"
	"github.com/itksb/go-url-shortener/internal/router"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/internal/storage"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/itksb/go-url-shortener/pkg/session"
	"io"
	"net/http"
	"time"
)

// App - application
type App struct {
	HTTPServer   *http.Server
	logger       logger.Interface
	urlshortener *shortener.Service
	io.Closer
}

// NewApp - constructor of the App
func NewApp(cfg config.Config) (*App, error) {
	l, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	var repo shortener.ShortenerStorage
	if cfg.FileStoragePath != "" {
		repo, err = filestorage.NewStorage(l, cfg.FileStoragePath)
		if err != nil {
			l.Error(fmt.Sprintf("File storage error: %s", err.Error()))
			return nil, err
		}
	} else {
		// inMemory storage
		repo = storage.NewStorage(l)
	}
	urlshortener := shortener.NewShortener(l, repo)
	h := handler.NewHandler(l, urlshortener, cfg)

	codec, err := session.NewSecureCookie([]byte(cfg.SessionConfig.HashKey), []byte(cfg.SessionConfig.BlockKey))
	if err != nil {
		l.Error(fmt.Sprintf("Codec for session creating error: %s", err.Error()))
		return nil, err
	}
	sessionStore := session.NewCookieStore(codec)

	routeHandler, err := router.NewRouter(h, sessionStore, l)
	if err != nil {
		l.Error(fmt.Sprintf("Router creating error: %s", err.Error()))
		return nil, err
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort),
		Handler:      routeHandler,
		WriteTimeout: 15 * time.Second,
	}

	return &App{
		HTTPServer:   srv,
		logger:       l,
		urlshortener: urlshortener,
	}, nil
}

// Run - run the application instance
func (app *App) Run() error {
	app.logger.Info("server started", "addr", app.HTTPServer.Addr)
	return app.HTTPServer.ListenAndServe()
}

// Close -
func (app *App) Close() error {
	return app.urlshortener.Close()
}
