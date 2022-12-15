package shortener

import (
	"context"
	"errors"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"io"
)

// ShortenerStorage -
//
//goland:noinspection GoNameStartsWithPackageName
type ShortenerStorage interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURL(ctx context.Context, id string) (string, error)
	io.Closer
}

// Service -
type Service struct {
	logger  logger.Interface
	storage ShortenerStorage
	io.Closer
}

// NewShortener - constructor
func NewShortener(l logger.Interface, storage ShortenerStorage) *Service {
	return &Service{
		logger:  l,
		storage: storage,
	}
}

// ShortenURL - saves the given url to the database and returns record id
func (s *Service) ShortenURL(ctx context.Context, url string) (string, error) {
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
		return "", errors.New("ShortenerStorage error: savedURL != url")
	}

	return id, nil
}

// GetURL - retrieves url by the id
func (s *Service) GetURL(ctx context.Context, id string) (string, error) {
	return s.storage.GetURL(ctx, id)
}

// Close -
func (s *Service) Close() error {
	return s.storage.Close()
}
