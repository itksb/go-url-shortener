package dbstorage

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestStorage_GetURL(t *testing.T) {
	// mock DB
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	l, err := logger.NewLogger()

	assert.NoError(t, err)
	// instantiate storage with mock DB
	storage, err := NewPostgres("dsn", l, db)
	assert.NoError(t, err)

	// test data
	testID := int64(123)
	expectedResult := shortener.URLListItem{
		ID:          testID,
		UserID:      "user1",
		OriginalURL: "https://www.example.com",
		DeletedAt:   nil,
	}

	// prepare mock DB expectations
	rows := sqlmock.NewRows([]string{"id", "user_id", "original_url", "deleted_at"}).
		AddRow(testID, "user1", "https://www.example.com", nil)
	mock.ExpectQuery("SELECT id, user_id, original_url, deleted_at FROM urls WHERE id = ?").
		WithArgs(testID).
		WillReturnRows(rows)

	// execute GetURL and check the result
	result, err := storage.GetURL(context.Background(), strconv.FormatInt(testID, 10))
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)

	// assert mock DB expectations
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)

}
