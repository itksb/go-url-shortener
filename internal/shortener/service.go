package shortener

import (
	"context"
	"errors"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"io"
)

//goland:noinspection GoNameStartsWithPackageName

// Service -
type Service struct {
	logger  logger.Interface
	storage ShortenerStorage
	io.Closer
}

// URLListItem - .
type URLListItem struct {
	ID          int64   `json:"-" db:"id"`
	UserID      string  `json:"-" db:"user_id"`
	ShortURL    string  `json:"short_url" db:"sql.Null*"`
	OriginalURL string  `json:"original_url" db:"original_url"`
	CreatedAt   string  `json:"-"  db:"created_at,sql.Null*"`
	UpdatedAt   string  `json:"-" db:"updated_at,sql.Null*"`
	DeletedAt   *string `json:"-" db:"deleted_at,sql.Null*"`
}

// InternalStats - .
type InternalStats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// NewShortener - constructor
func NewShortener(l logger.Interface, storage ShortenerStorage) *Service {
	return &Service{
		logger:  l,
		storage: storage,
	}
}

// ShortenURL - saves the given url to the database and returns record id
func (s *Service) ShortenURL(ctx context.Context, url string, userID string) (string, error) {
	if len(url) == 0 {
		return "", errors.New("empty url")
	}
	id, err := s.storage.SaveURL(ctx, url, userID)
	if err != nil && !errors.Is(err, ErrDuplicate) {
		return "", err
	}

	savedItem, err2 := s.storage.GetURL(ctx, id)
	if err2 != nil {
		return "", err
	}

	if savedItem.OriginalURL != url {
		return "", errors.New("ShortenerStorage error: savedItem != url")
	}

	return id, err
}

// GetURL - retrieves url by the id
func (s *Service) GetURL(ctx context.Context, id string) (URLListItem, error) {
	return s.storage.GetURL(ctx, id)
}

// ListURLByUserID - list urls shortened by the user
func (s *Service) ListURLByUserID(ctx context.Context, userID string) ([]URLListItem, error) {
	return s.storage.ListURLByUserID(ctx, userID)
}

// DeleteURLBatch - makrs urls
func (s *Service) DeleteURLBatch(ctx context.Context, userID string, ids []string) error {
	return s.storage.DeleteURLBatch(ctx, userID, ids)
}

// Close destructor
func (s *Service) Close() error {
	return s.storage.Close()
}
