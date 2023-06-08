package token

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestJwtMaker(t *testing.T) {
	secKey := util.RandomString(32)
	jmaker, err := NewJwtMaker(secKey)
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute

	issueAt := time.Now()
	expireAt := time.Now().Add(duration)

	token, payload, err := jmaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = jmaker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.Equal(t, payload.Username, username)
	require.NotZero(t, payload.Id)
	require.WithinDuration(t, payload.IssueAt, issueAt, time.Second)
	require.WithinDuration(t, payload.ExpireAt, expireAt, time.Second)
}

func TestExpiredJwtMaker(t *testing.T) {
	secKey := util.RandomString(32)
	jmaker, err := NewJwtMaker(secKey)
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := -time.Minute

	token, payload, err := jmaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = jmaker.VerifyToken(token)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrExpireToken))
	require.Nil(t, payload)
}

func TestInvalidJwtMaker(t *testing.T) {
	secKey := util.RandomString(32)
	jmaker, err := NewJwtMaker(secKey)
	require.NoError(t, err)

	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	require.NotEmpty(t, token)

	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	p, err := jmaker.VerifyToken(tokenStr)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, p)
}
