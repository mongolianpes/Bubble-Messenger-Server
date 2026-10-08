package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	usersServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	db := setupTestDB(t)

	ctx := context.Background()

	registerUser1 := "user1"
	registerUser := "user"
	registeredUserAsLogin := "login"
	_, err = db.Exec("INSERT INTO users (login, name, password) VALUES ($1, $2, $3)", registeredUserAsLogin, "login", "password123")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO users (login, name, password) VALUES ($1, $2, $3)", registerUser, "user", "password123")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO users (login, name, password) VALUES ($1, $2, $3)", registerUser1, "user1", "password123")
	require.NoError(t, err)

	findUsers, err := client.Search(ctx, "user")
	require.NoError(t, err)

	for _, user := range findUsers {
		require.Contains(t, []string{registerUser, registerUser1}, user)
	}
}

func TestGetInfoByID(t *testing.T) {
	usersServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	db := setupTestDB(t)

	ctx := context.Background()

	registerUser1 := "user1"
	registerUser := "user"
	registeredUserAsLogin := "login"
	_, err = db.Exec("INSERT INTO users (login, name, password) VALUES ($1, $2, $3)", registeredUserAsLogin, "user", "password123")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO users (login, name, password) VALUES ($1, $2, $3)", registerUser, "user", "password123")
	require.NoError(t, err)

	var userID int
	err = db.QueryRow("INSERT INTO users (login, name, password) VALUES ($1, $2, $3) RETURNING id", registerUser1, "user", "password123").Scan(&userID)
	require.NoError(t, err)

	findUser, err := client.GetInfoByID(ctx, int(userID))
	require.NoError(t, err)
	require.Equal(t, findUser.Login, registerUser1)
	require.Equal(t, findUser.Name, "user")
}
