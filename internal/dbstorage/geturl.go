// Package dbstorage used for persisting urls in the database
package dbstorage

import (
	"context"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"strconv"
)

// GetURL retrieves url from the underlying db by id
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
	err = res.Scan(&result.ID, &result.UserID, &result.OriginalURL, &result.DeletedAt)
	if err != nil {
		s.l.Error(err)
		return result, err
	}

	return result, nil
}
