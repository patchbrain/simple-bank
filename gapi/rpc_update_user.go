package gapi

import (
	"context"
	"database/sql"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/pb"
	"github.com/patchbrain/simple-bank/util"
	"github.com/patchbrain/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// 增加限制，只有当前用户能够更改自己的信息
	// 验证令牌有效性
	userPayload, err := s.AuthorizeUser(ctx)
	if err != nil {
		return nil, authorizationError(err)
	}

	// 验证用户一致性
	if userPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot modify other user's info")
	}

	// 验证参数，并对错误列表进行处理
	violations := ValidateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	params := db.UpdateUserParams{}
	if req.GetPassword() != "" {
		// 要更新密码
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "fail to hash password: %s", err)
		}

		params.PasswordHashed = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}
		// 同时更新 ”密码更新时间“
		params.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	if req.GetEmail() != "" {
		params.Email = sql.NullString{
			String: req.GetEmail(),
			Valid:  true,
		}
	}

	if req.GetFullName() != "" {
		params.FullName = sql.NullString{
			String: req.GetFullName(),
			Valid:  true,
		}
	}

	params.Username = req.GetUsername()
	user, err := s.Store.UpdateUser(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "cannot find user: %s", params.Username)
		}
		return nil, status.Errorf(codes.Internal, "fail to update user: %s", err)
	}

	rsp := &pb.UpdateUserResponse{
		User: convertUser(user),
	}

	return rsp, nil
}

func ValidateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	var err error
	username := req.GetUsername()
	val.ValidateUsername(username, 3, 100)
	if err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.GetFullName() != "" {
		err = val.ValidateFullName(req.GetFullName(), 3, 100)
		if err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if req.GetPassword() != "" {
		err = val.ValidatePassword(req.GetPassword())
		if err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if req.GetEmail() != "" {
		err = val.ValidateEmail(req.GetEmail())
		if err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return violations
}
