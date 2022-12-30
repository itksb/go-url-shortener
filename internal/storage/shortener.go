package storage

import (
	"context"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"strconv"
)

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

func (s *storage) GetURL(ctx context.Context, id string) (string, error) {
	s.urlMtx.RLock()
	defer s.urlMtx.RUnlock()

	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return "", err
	}

	urlListItem, ok := s.urls[idInt64]
	if !ok {
		return "", fmt.Errorf("urlListItem with id %d is not exists", idInt64)
	}

	return urlListItem.OriginalURL, nil
}

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

func (s *storage) Close() error { return nil }
