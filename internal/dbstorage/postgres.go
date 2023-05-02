// Package dbstorage used for persisting urls in the database
package dbstorage

import (
	"context"
	"database/sql"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/jmoiron/sqlx"
	//Under the hood, the driver registers itself as being available to the database/sql package,
	//but in general nothing else happens with the exception that the init function is run.
	_ "github.com/lib/pq"
)

// Storage database service
// implements ShortenerStorage interface and Closer interface
type Storage struct {
	dsn string
	db  *sqlx.DB
	l   logger.Interface
}

const dbDriverName = "postgres"

// NewPostgres - postgres service constructor
// sqlDB can be nil. If nil, then it will be created
func NewPostgres(dsn string, l logger.Interface, sqlDB *sql.DB) (*Storage, error) {
	var db *sqlx.DB
	var err error
	if sqlDB != nil {
		db = sqlx.NewDb(sqlDB, dbDriverName)
	} else {
		db, err = sqlx.Connect(dbDriverName, dsn)
	}

	return &Storage{
		dsn: dsn,
		db:  db,
		l:   l,
	}, err
}

// reconnect to db if connection is not created
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

// Ping check whether connection to db is valid or not
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

// Close destructor
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
