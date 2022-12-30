package shortener

import (
	"context"
	"errors"
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

// ErrDuplicate - duplication error returns from the storage
var ErrDuplicate = errors.New(`duplicate entity`)
