package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"strconv"
	"time"
)

// SaveURL persist the given url
func (s *storage) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	s.urlMtx.Lock()
	defer s.urlMtx.Unlock()

	id := s.currentURLID
	s.currentURLID++
	if _, ok := s.urls[id]; ok {
		return "0", fmt.Errorf("url with id %d already exists", id)
	}
	s.urls[id] = shortener.URLListItem{
		ID:          id,
		UserID:      userID,
		OriginalURL: url,
	}
	return fmt.Sprint(id), nil
}

// GetURL retrieve url
func (s *storage) GetURL(ctx context.Context, id string) (shortener.URLListItem, error) {
	s.urlMtx.RLock()
	defer s.urlMtx.RUnlock()

	result := shortener.URLListItem{}

	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return result, err
	}

	urlListItem, ok := s.urls[idInt64]
	if !ok {
		return result, fmt.Errorf("urlListItem with id %d is not exists", idInt64)
	}

	return urlListItem, nil
}

// ListURLByUserID returns the list of urls
func (s *storage) ListURLByUserID(ctx context.Context, userID string) ([]shortener.URLListItem, error) {
	s.urlMtx.RLock()
	defer s.urlMtx.RUnlock()
	var items []shortener.URLListItem

	for _, item := range s.urls {
		if item.UserID == userID {
			items = append(items, item)
		}
	}

	return items, nil
}

// DeleteURLBatch removes urls
func (s *storage) DeleteURLBatch(ctx context.Context, userID string, ids []string) error {
	s.urlMtx.Lock()
	defer s.urlMtx.Unlock()
	var hasError bool
	for i := 0; i < len(ids); i++ {
		idInt64, err := strconv.ParseInt(ids[i], 10, 64)
		if err != nil {
			return err
		}

		// get a "copy" here
		if entry, ok := s.urls[idInt64]; ok {
			if entry.UserID == userID {
				tCurr := time.Now().Format("2006-01-02T15:04:05")
				entry.DeletedAt = &tCurr
				s.urls[idInt64] = entry
			}
		} else {
			hasError = true
		}
	}
	if hasError {
		return errors.New("some keys not used")
	}
	return nil
}

// Close destructor
func (s *storage) Close() error { return nil }
