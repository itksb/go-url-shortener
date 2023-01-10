package handler

import (
	"context"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"strconv"
)

type storageMock struct {
	urls         map[int64]shortener.URLListItem
	currentURLID int64
}

func newStorageMock(urls map[int64]shortener.URLListItem) *storageMock {
	return &storageMock{urls: urls, currentURLID: 0}
}

func (s *storageMock) SaveURL(ctx context.Context, url string, userID string) (string, error) {
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

func (s *storageMock) GetURL(ctx context.Context, id string) (shortener.URLListItem, error) {
	idInt64, err := strconv.ParseInt(id, 10, 64)
	item := shortener.URLListItem{}
	if err != nil {
		return item, err
	}

	url, ok := s.urls[idInt64]
	if !ok {
		return item, fmt.Errorf("url with id %d is not exists", idInt64)
	}

	return url, nil
}

func (s *storageMock) ListURLByUserID(ctx context.Context, userID string) ([]shortener.URLListItem, error) {
	var items []shortener.URLListItem

	for _, item := range s.urls {
		if item.UserID == userID {
			items = append(items, item)
		}
	}

	return items, nil
}

func (s *storageMock) DeleteURLBatch(ctx context.Context, userID string, ids []string) error {
	return nil
}

func (s *storageMock) Close() error { return nil }
