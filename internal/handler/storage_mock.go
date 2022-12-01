package handler

import (
	"context"
	"fmt"
	"strconv"
)

type storageMock struct {
	urls         map[int64]string
	currentURLID int64
}

func newStorageMock(urls map[int64]string) *storageMock {
	return &storageMock{urls: urls, currentURLID: 0}
}

func (s *storageMock) SaveURL(ctx context.Context, url string) (string, error) {
	id := s.currentURLID
	s.currentURLID++
	if _, ok := s.urls[id]; ok {
		return "0", fmt.Errorf("url with id %d already exists", id)
	}
	s.urls[id] = url
	return fmt.Sprint(id), nil
}

func (s *storageMock) GetURL(ctx context.Context, id string) (string, error) {
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return "", err
	}

	url, ok := s.urls[idInt64]
	if !ok {
		return "", fmt.Errorf("url with id %d is not exists", idInt64)
	}

	return url, nil
}
