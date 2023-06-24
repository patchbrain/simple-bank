package db

import (
	"context"
)

type TxCreateUserParam struct {
	CreateUserParams
	AfterCreate func(user User) error
}

type TxCreateUserResult struct {
	User User
}

func (s *SQLStore) TxCreateUser(ctx context.Context, param TxCreateUserParam) (TxCreateUserResult, error) {
	var result TxCreateUserResult

	err := s.txExec(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, param.CreateUserParams)

		if err != nil {
			return err
		}

		return param.AfterCreate(result.User)
	})

	return result, err
}
