package gapi

import (
	"context"
	"database/sql"
	"errors"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/pb"
	"github.com/patchbrain/simple-bank/util"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	user, err := s.Store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "fail to get user: %s", err)
		}

		return nil, status.Errorf(codes.Internal, "fail to get user: %s", err)
	}

	err = util.CheckPassword(req.GetPassword(), user.PasswordHashed)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, status.Errorf(codes.Unauthenticated, "error password: %s", err)
		}

		return nil, status.Errorf(codes.Internal, "fail to verify password: %s", err)
	}

	token, payload, err := s.TokenMaker.CreateToken(user.Username, s.Config.TokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to create token: %s", err)
	}

	// 再创建一个refreshtoken
	refreshToken, refreshPayload, err := s.TokenMaker.CreateToken(user.Username, s.Config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to create refresh token: %s", err)
	}

	// 存入数据库
	session, err := s.Store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.Id,
		Username:     refreshPayload.Username,
		RefreshToken: refreshToken,
		IsBlocked:    false,
		ExpiredAt:    refreshPayload.ExpireAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to create session: %s", err)
	}

	rsp := new(pb.LoginUserResponse)
	rsp.User = convertUser(user)
	rsp.SessionId = session.ID.String()
	rsp.AccessToken = token
	rsp.AccessTokenExpiredAt = timestamppb.New(payload.ExpireAt)
	rsp.RefreshToken = refreshToken
	rsp.RefreshTokenExpiredAt = timestamppb.New(refreshPayload.ExpireAt)

	return rsp, nil
}
