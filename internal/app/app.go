package app

import (
	"fmt"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/handler"
	"github.com/itksb/go-url-shortener/internal/router"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/internal/storage"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"net/http"
)

// App - application
type App struct {
	HTTPServer *http.Server
	logger     logger.Interface
}

// NewApp - constructor of the App
func NewApp(cfg config.Config) (*App, error) {
	l, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	repo := storage.NewStorage(l)
	urlshortener := shortener.NewShortener(l, repo)
	h := handler.NewHandler(l, urlshortener, cfg)

	router := router.NewRouter(h)

	srv := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", cfg.AppPort),
	}

	return &App{
		HTTPServer: srv,
		logger:     l,
	}, nil
}

// Run - run the application instance
func (app *App) Run() error {
	app.logger.Info("server started", "addr", app.HTTPServer.Addr)
	return app.HTTPServer.ListenAndServe()
}
