package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/patchbrain/simple-bank/token"
	"net/http"
	"strings"
)

const (
	authHeaderKey  = "authorization"
	authTypeBearer = "bearer"
	authPayloadKey = "auth_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headerValue := ctx.GetHeader(authHeaderKey)
		if headerValue == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("no authorization headerValue")))
			return
		}

		parts := strings.Fields(headerValue)
		if len(parts) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("wrong authorization format")))
			return
		}

		if parts[0] != authTypeBearer {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("unsupported authorization type")))
			return
		}

		token := parts[1]
		payload, err := tokenMaker.VerifyToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authPayloadKey, payload)
		ctx.Next()
	}
}
