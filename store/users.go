package store

import (
	"context"

	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
)

// CreateUserTxParams stores input params of the transfer transaction.
type CreateUserTxParams struct {
	simplebanksql.CreateUserParams
	// AfterCreat is callback to execute when a user was already created.
	AfterCreat func(user simplebanksql.User) error
}

// CreateUserTxResult stores the result of a transfer transaction.
type CreateUserTxResult struct {
	User simplebanksql.User
}

func (s *SimpleBankDB) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult
	err := s.execWithContext(ctx, func(q *simplebanksql.Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}
		return arg.AfterCreat(result.User)
	})

	return result, err
}
