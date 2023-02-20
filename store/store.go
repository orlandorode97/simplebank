package store

import (
	"context"
	"database/sql"

	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
)

// Stores provides all needed functional sql database
type Store interface {
	Ping() error
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	simplebanksql.Querier
}

// SimpleBankDB stores the primary database and sqlc generated code.
type SimpleBankDB struct {
	db *sql.DB
	*simplebanksql.Queries
}

// NewSimpleBankDB returns a *SimpleBankDB.
func NewSimpleBankDB(db *sql.DB) Store {
	return &SimpleBankDB{
		db:      db,
		Queries: simplebanksql.New(db),
	}
}

func (s *SimpleBankDB) Ping() error {
	return s.db.Ping()
}
