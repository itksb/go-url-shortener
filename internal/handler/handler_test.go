package handler

import (
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/dbstorage"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHandler(t *testing.T) {
	l := &logger.Logger{}
	s := &shortener.Service{}
	db := &dbstorage.Storage{}
	dbping := db
	cfg := config.Config{}

	h := NewHandler(l, s, db, dbping, cfg)
	assert.NotNil(t, h)
}
