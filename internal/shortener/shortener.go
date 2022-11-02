package shortener

import (
	"context"
	"errors"
	"github.com/itksb/go-url-shortener/pkg/logger"
)

type storage interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURL(ctx context.Context, id string) (string, error)
}

type service struct {
	logger  logger.Interface
	storage storage
}

// NewShortener - constructor
func NewShortener(l logger.Interface, storage storage) *service {
	return &service{
		logger:  l,
		storage: storage,
	}
}

// ShortenUrl - saves the given url to the database and returns record id
func (s *service) ShortenURL(ctx context.Context, url string) (string, error) {
	if len(url) == 0 {
		return "", errors.New("empty url")
	}
	id, err := s.storage.SaveURL(ctx, url)
	if err != nil {
		return "", err
	}

	savedURL, err := s.storage.GetURL(ctx, id)
	if err != nil {
		return "", err
	}

	if savedURL != url {
		return "", errors.New("storage error")
	}

	return id, nil
}

// GetUrl - retreives url by the id
func (s *service) GetURL(ctx context.Context, id string) (string, error) {
	return s.storage.GetURL(ctx, id)
}
