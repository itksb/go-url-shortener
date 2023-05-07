package handler

import (
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/dbstorage"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/pkg/logger"
)

// Handler - endpoint handlers
type Handler struct {
	logger       logger.Interface
	urlshortener *shortener.Service
	cfg          config.Config
	dbservice    *dbstorage.Storage
	dbping       IPingableDB
}

// NewHandler - constructor
func NewHandler(
	logger logger.Interface,
	shortener *shortener.Service,
	dbservice *dbstorage.Storage,
	dbping IPingableDB,
	cfg config.Config,
) *Handler {
	return &Handler{
		logger:       logger,
		urlshortener: shortener,
		cfg:          cfg,
		dbservice:    dbservice,
		dbping:       dbping,
	}
}
