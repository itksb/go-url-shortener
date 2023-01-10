package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"strconv"
)

// SaveURL - saves url to the postgres db
func (s *Storage) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return "", err
	}

	query := `INSERT INTO urls (user_id, original_url) VALUES ($1, $2)
              ON CONFLICT ON CONSTRAINT urls_unique_idx DO NOTHING RETURNING id`
	row := s.db.QueryRowContext(ctx, query, userID, url)

	var ID int
	err = row.Scan(&ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//query does not return id, so duplicate conflict, need to retrieve id from db
			row := s.db.QueryRowContext(ctx, `SELECT id FROM urls WHERE original_url = $1`, url)
			err = row.Scan(&ID)
			if err != nil {
				s.l.Error(err)
				return "", err
			}
			return fmt.Sprint(ID), fmt.Errorf("%w", shortener.ErrDuplicate)
		}

	}

	return fmt.Sprint(ID), nil
}

// GetURL - retrieves url from the underlying db by id
func (s *Storage) GetURL(ctx context.Context, id string) (shortener.URLListItem, error) {
	result := shortener.URLListItem{}
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return result, err
	}

	query := `SELECT id, user_id, original_url, deleted_at FROM urls WHERE id = $1`

	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return result, err
	}
	res := s.db.QueryRowContext(ctx, query, idInt64)

	err = res.Scan(&result)
	if err != nil {
		s.l.Error(err)
		return result, err
	}

	return result, nil
}

// ListURLByUserID - list urls by user
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

func (s *Storage) DeleteURLBatch(ctx context.Context, userID string, ids []string) error {
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return "", err
	}

}
