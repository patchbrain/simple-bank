package gapi

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/pb"
	"github.com/patchbrain/simple-bank/util"
	"github.com/patchbrain/simple-bank/val"
	"github.com/patchbrain/simple-bank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// 验证参数，并对错误列表进行处理
	violations := ValidateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to hash password: %s", err)
	}

	params := db.TxCreateUserParam{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			PasswordHashed: hashedPassword,
			FullName:       req.GetFullName(),
			Email:          req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(5 * time.Second),
				asynq.Group(worker.QueueNameCritical),
			}
			payload := worker.VerifyEmailPayload{Username: req.GetUsername()}
			return s.TaskDistributor.EnqueueTask(ctx, payload, opts...)
		},
	}

	result, err := s.Store.TxCreateUser(ctx, params)
	if err != nil {
		if pqerr, ok := err.(*pq.Error); ok {
			switch pqerr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "fail to create User: %s", err)
			}
		}

		return nil, status.Errorf(codes.Internal, "fail to create User: %s", err)
	}

	rsp := &pb.CreateUserResponse{
		User: convertUser(result.User),
	}

	return rsp, nil
}

func ValidateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	err := val.ValidateUsername(req.GetUsername(), 3, 100)
	if err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	err = val.ValidateFullName(req.GetFullName(), 3, 100)
	if err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	err = val.ValidatePassword(req.GetPassword())
	if err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	err = val.ValidateEmail(req.GetEmail())
	if err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}
