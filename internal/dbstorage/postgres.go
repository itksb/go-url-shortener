package dbstorage

import (
	"context"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/jmoiron/sqlx"
	//Under the hood, the driver registers itself as being available to the database/sql package,
	//but in general nothing else happens with the exception that the init function is run.
	_ "github.com/lib/pq"
)

// Storage - abstract database service
type Storage struct {
	db *sqlx.DB
	l  logger.Interface
}

// NewPostgres - postgres service constructor
func NewPostgres(dsn string, l logger.Interface) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
		l:  l,
	}, nil
}

// Ping - check whether connection to db is valid or not
func (s *Storage) Ping(ctx context.Context) bool {
	err := s.db.PingContext(ctx)
	if err != nil {
		s.l.Error(err)
		return false
	}
	return true
}

// Close - destructor
func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close()
}
