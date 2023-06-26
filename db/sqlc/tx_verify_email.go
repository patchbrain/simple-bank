package db

import (
	"context"
	"database/sql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TxVerifyEmailParam struct {
	Id         int64  `json:"id"`
	SecretCode string `json:"secret_code"`
}

type TxVerifyEmailResult struct {
	IsVerified bool `json:"is_verified"`
}

func (s *SQLStore) TxVerifyEmail(ctx context.Context, param TxVerifyEmailParam) (TxVerifyEmailResult, error) {
	var result TxVerifyEmailResult

	err := s.txExec(ctx, func(q *Queries) error {
		verify, err := q.GetEmailVerify(ctx, param.Id)
		if err != nil {
			return status.Errorf(codes.Internal, "fail to get email verify object: %s", err.Error())
		}
		// 该用户是否已验证
		user, err := q.GetUser(ctx, verify.Username)
		if user.IsVerified {
			return status.Errorf(codes.Unauthenticated, "email verification cannot be repeated: %s", err.Error())
		}

		// 更新verify字段
		_, err = q.UpdateEmailVerify(ctx, UpdateEmailVerifyParams{
			SecretCode: param.SecretCode,
			ID:         param.Id,
		})
		if err != nil {
			return status.Errorf(codes.Internal, "fail to update email verify info: %s", err.Error())
		}

		// 验证成功则修改user的验证字段
		params := UpdateUserParams{
			Username: user.Username,
			IsVerified: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		}

		_, err = q.UpdateUser(ctx, params)
		if err != nil {
			if err == sql.ErrNoRows {
				return status.Errorf(codes.NotFound, "cannot find user: %s", params.Username)
			}
			return status.Errorf(codes.Internal, "fail to update user: %s", err.Error())
		}

		result.IsVerified = true
		return nil
	})

	return result, err
}
