package dbstorage

import (
	"context"
	"github.com/itksb/go-url-shortener/internal/shortener"
)

// GetURL retrieves url from the underlying db by id
func (s *Storage) GetStats(ctx context.Context) (shortener.InternalStats, error) {
	result := shortener.InternalStats{}
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return result, err
	}

	query := `
SELECT
    COUNT(DISTINCT user_id) AS user_count,
    COUNT(original_url) AS url_count
FROM urls
WHERE deleted_at IS NULL`
	res := s.db.QueryRowContext(ctx, query)
	err = res.Scan(&result.Users, &result.URLs)
	if err != nil {
		s.l.Error(err)
		return result, err
	}

	return result, nil
}
