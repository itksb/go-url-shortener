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
	dsn string
	db  *sqlx.DB
	l   logger.Interface
}

// NewPostgres - postgres service constructor
func NewPostgres(dsn string, l logger.Interface) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	return &Storage{
		dsn: dsn,
		db:  db,
		l:   l,
	}, err
}

func (s *Storage) reconnect(ctx context.Context) error {
	var err error
	if s.db == nil {
		s.db, err = sqlx.ConnectContext(ctx, "postgres", s.dsn)
		if err != nil {
			return err
		}
	}
	return nil
}

// Ping - check whether connection to db is valid or not
func (s *Storage) Ping(ctx context.Context) bool {
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return false
	}
	err = s.db.PingContext(ctx)
	if err != nil {
		s.l.Error(err)
		return false
	}
	return true
}

// Close - destructor
func (s *Storage) Close() error {
	return s.db.Close()
}
