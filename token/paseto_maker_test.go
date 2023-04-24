package token

import (
	"errors"
	"github.com/o1egl/paseto"
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPasetoMaker(t *testing.T) {
	secKey := util.RandomString(32)
	maker, err := NewPasetoMaker(secKey)
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute

	issueAt := time.Now()
	expireAt := time.Now().Add(duration)

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.Equal(t, payload.Username, username)
	require.NotZero(t, payload.Id)
	require.WithinDuration(t, payload.IssueAt, issueAt, time.Second)
	require.WithinDuration(t, payload.ExpireAt, expireAt, time.Second)
}

func TestExpiredPasetoMaker(t *testing.T) {
	secKey := util.RandomString(32)
	maker, err := NewJwtMaker(secKey)
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := -time.Minute

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrExpireToken))
	require.Nil(t, payload)
}

func TestInvalidPasetoMaker(t *testing.T) {
	secKey := util.RandomString(32)
	maker, err := NewPasetoMaker(secKey)
	require.NoError(t, err)

	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	p := paseto.NewV2()
	tokenStr, err := p.Encrypt([]byte(secKey), payload, nil)
	require.NoError(t, err)

	tokenStr = "v9.local." + tokenStr[8:]
	payload, err = maker.VerifyToken(tokenStr)
	require.Nil(t, payload)
	require.EqualError(t, err, ErrInvalidToken.Error())
}
