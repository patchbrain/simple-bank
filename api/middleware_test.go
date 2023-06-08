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
	token, payload, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

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
		{
			name: "No AuthHeader",
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)
			},
		},
		{
			name: "Unsupported Auth Type",
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)
			},
		},
		{
			name: "Error Format",
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)
			},
		},
		{
			name: "Expired AccessToken",
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)
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
