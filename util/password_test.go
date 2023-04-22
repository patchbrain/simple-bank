package util

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestPassword(t *testing.T) {
	password := RandomString(8)
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)

	err = CheckPassword(password, hashedPassword)
	require.NoError(t, err)

	passwordError := RandomString(8)
	err = CheckPassword(passwordError, hashedPassword)
	require.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword)
}
