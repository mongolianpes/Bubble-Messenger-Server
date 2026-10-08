package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetAndGetAvatar(t *testing.T) {
	usersServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	db := setupTestDB(t)

	ctx := context.Background()

	login := "user1"
	name := "user1"
	password := "password123"
	_, err = db.Exec("INSERT INTO users (login, name, password) VALUES ($1, $2, $3)", login, name, password)
	require.NoError(t, err)

	avatarPath := "avatar-path"
	err = client.SetAvatar(ctx, login, avatarPath)
	require.NoError(t, err)

	gettedAvatarPath, err := client.GetAvatar(ctx, login)
	require.NoError(t, err)
	require.Equal(t, gettedAvatarPath, avatarPath)
}
