package dbstorage

import (
	"context"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"strconv"
)

func (s *Storage) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return "", err
	}

	query := `INSERT INTO urls (user_id, original_url) VALUES ($1, $2) RETURNING id`
	row := s.db.QueryRowContext(ctx, query, userID, url)
	if err != nil {
		s.l.Error(err)
		return "", err
	}
	var ID int
	err = row.Scan(&ID)

	if err != nil {
		s.l.Error(err)
		return "", err
	}

	return fmt.Sprint(ID), nil
}

func (s *Storage) GetURL(ctx context.Context, id string) (string, error) {
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return "", err
	}

	query := `SELECT original_url FROM urls WHERE id = $1`

	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return "", err
	}
	res := s.db.QueryRowContext(ctx, query, idInt64)

	var originalURL string
	err = res.Scan(&originalURL)
	if err != nil {
		s.l.Error(err)
		return "", err
	}

	return originalURL, nil
}

func (s *Storage) ListURLByUserID(ctx context.Context, userID string) ([]shortener.URLListItem, error) {
	urls := []shortener.URLListItem{}
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return urls, err
	}

	query := "SELECT * FROM urls WHERE user_id=$1"

	err = s.db.Select(&urls, query, userID)
	if err != nil {
		s.l.Error(err)
		return urls, err
	}

	return urls, nil

}
