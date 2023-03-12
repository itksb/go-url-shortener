// Package dbstorage used for persisting urls in the database
package dbstorage

import (
	"context"
	"github.com/itksb/go-url-shortener/internal/shortener"
)

// ListURLByUserID list urls by user
func (s *Storage) ListURLByUserID(ctx context.Context, userID string) ([]shortener.URLListItem, error) {
	var urls = []shortener.URLListItem{}
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
