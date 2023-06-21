package db

import (
	"context"
	"database/sql"
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	password := util.RandomString(8)
	hashed, err := util.HashPassword(password)
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		PasswordHashed: hashed,
		FullName:       util.RandomString(6),
		Email:          util.RandomEmail(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, user)

	require.Equal(t, user.Username, arg.Username)
	require.Equal(t, user.PasswordHashed, arg.PasswordHashed)
	require.Equal(t, user.FullName, arg.FullName)
	require.Equal(t, user.Email, arg.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)

	require.NoError(t, err)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.PasswordHashed, user2.PasswordHashed)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)

	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
}

func TestUpdateUserOnlyForFullName(t *testing.T) {
	oldUser := createRandomUser(t)
	newFullName := util.RandomOwner()
	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		FullName: sql.NullString{String: newFullName, Valid: true},
		Username: oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.FullName, newUser.FullName)
	require.Equal(t, oldUser.Email, newUser.Email)
	require.Equal(t, oldUser.Username, newUser.Username)
	require.Equal(t, oldUser.PasswordHashed, newUser.PasswordHashed)
}

func TestUpdateUserOnlyForEmail(t *testing.T) {
	oldUser := createRandomUser(t)
	newEmail := util.RandomEmail()
	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Email:    sql.NullString{String: newEmail, Valid: true},
		Username: oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Email, newUser.Email)
	require.Equal(t, oldUser.FullName, newUser.FullName)
	require.Equal(t, oldUser.Username, newUser.Username)
	require.Equal(t, oldUser.PasswordHashed, newUser.PasswordHashed)
}

func TestUpdateUserOnlyForPasswordHashed(t *testing.T) {
	oldUser := createRandomUser(t)
	newPasswordHashed, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		PasswordHashed: sql.NullString{String: newPasswordHashed, Valid: true},
		Username:       oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.PasswordHashed, newUser.PasswordHashed)
	require.Equal(t, oldUser.FullName, newUser.FullName)
	require.Equal(t, oldUser.Username, newUser.Username)
	require.Equal(t, oldUser.Email, newUser.Email)
}
