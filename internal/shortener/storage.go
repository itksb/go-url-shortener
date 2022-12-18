package shortener

import (
	"context"
	"io"
)

// ShortenerStorage -
//
//goland:noinspection GoNameStartsWithPackageName
type ShortenerStorage interface {
	SaveURL(ctx context.Context, url string, userID string) (string, error)
	GetURL(ctx context.Context, id string) (string, error)
	ListURLByUserID(ctx context.Context, userID string) ([]URLListItem, error)
	io.Closer
}
