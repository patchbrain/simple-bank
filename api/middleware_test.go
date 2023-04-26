package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/patchbrain/simple-bank/token"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func addAuthHeader(t *testing.T, maker token.Maker, r *http.Request, authHeaderType string, username string, duration time.Duration) {
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	r.Header.Set(authHeaderKey, fmt.Sprintf("%s %s", authHeaderType, token))
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name            string
		setupAuthHeader func(t *testing.T, r *http.Request, maker token.Maker)
		checkResponse   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusOK)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(t, nil)

			uri := "/auth"
			s.Router.GET(uri, authMiddleware(s.TokenMaker), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			})

			request, err := http.NewRequest(http.MethodGet, uri, nil)
			require.NoError(t, err)

			tc.setupAuthHeader(t, request, s.TokenMaker)

			recorder := httptest.NewRecorder()

			s.Router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
