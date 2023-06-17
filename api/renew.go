package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type RenewTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RenewTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) renewToken(ctx *gin.Context) {
	var req RenewTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 检查token是否有效
	refreshPayload, err := s.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	session, err := s.Store.GetSession(ctx, refreshPayload.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 检查是否被blocked
	if session.IsBlocked == true {
		err = fmt.Errorf("blocked session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// 检查是否过期
	if time.Now().After(session.ExpiredAt) {
		err = fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if session.Username != refreshPayload.Username {
		err = fmt.Errorf("mismatched username")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		err = fmt.Errorf("error renewToken")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	token, payload, err := s.TokenMaker.CreateToken(session.Username, s.Config.TokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var rsp RenewTokenResponse
	rsp.AccessToken = token
	rsp.AccessTokenExpiresAt = payload.ExpireAt

	ctx.JSON(http.StatusOK, rsp)
}
