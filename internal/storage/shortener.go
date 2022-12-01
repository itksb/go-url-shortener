package storage

import (
	"context"
	"fmt"
	"strconv"
)

func (s *storage) SaveURL(ctx context.Context, url string) (string, error) {
	s.urlMtx.Lock()
	defer s.urlMtx.Unlock()

	id := s.currentURLID
	s.currentURLID++
	if _, ok := s.urls[id]; ok {
		return "0", fmt.Errorf("url with id %d already exists", id)
	}
	s.urls[id] = url
	return fmt.Sprint(id), nil
}

func (s *storage) GetURL(ctx context.Context, id string) (string, error) {
	s.urlMtx.RLock()
	defer s.urlMtx.RUnlock()

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
