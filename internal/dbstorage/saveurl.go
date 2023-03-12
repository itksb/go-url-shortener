package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
)

// SaveURL persist url
// to the database
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
