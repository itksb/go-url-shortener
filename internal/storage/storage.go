package storage

import (
	"github.com/itksb/go-url-shortener/pkg/logger"
	"sync"
)

type storage struct {
	logger logger.Interface
	urls   map[int64]string

	currentURLID int64
	urlMtx       sync.RWMutex
}

// NewStorage - constructor
func NewStorage(logger logger.Interface) *storage {
	return &storage{
		logger: logger,
		urls:   make(map[int64]string),
	}
}
