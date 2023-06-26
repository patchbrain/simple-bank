package gapi

import (
	"context"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/pb"
	"github.com/patchbrain/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func (s *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	id := req.GetId()
	secretCode := req.GetSecretCode()

	// 验证参数，并对错误列表进行处理
	violations := ValidateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	param := db.TxVerifyEmailParam{
		Id:         id,
		SecretCode: secretCode,
	}

	result, err := s.Store.TxVerifyEmail(ctx, param)
	if err != nil {
		return nil, err
	}

	rsp := &pb.VerifyEmailResponse{
		IsVerified: result.IsVerified,
	}
	return rsp, nil
}

func ValidateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateSecretCode(req.GetSecretCode(), 32); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}

	if err := val.ValidateEmailId(req.GetId()); err != nil {
		violations = append(violations, fieldViolation("id", err))
	}

	return violations
}
