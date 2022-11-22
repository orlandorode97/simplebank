package postgres

import (
	"database/sql"

	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
)

// SimpleBankDB stores the primary database and sqlc generated code.
type SimpleBankDB struct {
	db      *sql.DB
	queries *simplebanksql.Queries
}

// NewSimpleBankDB returns a *SimpleBankDB.
func NewSimpleBankDB(db *sql.DB) *SimpleBankDB {
	return &SimpleBankDB{
		db:      db,
		queries: simplebanksql.New(db),
	}
}
