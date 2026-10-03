package messenger

import (
	"bubble/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

func getTestDSN() string {
	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "5431")
	user := envOrDefault("TEST_DB_USER", "postgres")
	password := envOrDefault("TEST_DB_PASSWORD", "123")
	dbname := envOrDefault("TEST_DB_NAME", "bubble")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setupTestDB(t *testing.T) {
	t.Helper()

	db, err := sql.Open("postgres", getTestDSN())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, db.PingContext(ctx))

	_, err = db.ExecContext(ctx, `TRUNCATE TABLE messages RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func TestMessenger(t *testing.T) {
	messengerServiceHost = "localhost:8086"
	setupTestDB(t)

	ctx := context.Background()
	mockUsersService := NewUsersMockService()

	firstUserLogin := "user1"
	secondUserLogin := "user2"
	mockUsersService.On("GetInfoByLogin", ctx, firstUserLogin).Return(models.FindUser{
		Login: firstUserLogin,
		Name:  "User 1",
		ID:    1,
	}, nil)

	mockUsersService.On("GetInfoByLogin", ctx, secondUserLogin).Return(models.FindUser{
		Login: secondUserLogin,
		Name:  "User 2",
		ID:    2,
	}, nil)
	mockUsersService.On("GetInfoByID", ctx, 1).Return(models.FindUser{
		Login: firstUserLogin,
		Name:  "User 1",
		ID:    1,
	}, nil)

	messengerService, err := NewClient()
	require.NoError(t, err)

	fileName := "test.txt"
	storageFilePath := "path/test.txt"
	err = messengerService.SendFile(ctx, mockUsersService, firstUserLogin, secondUserLogin, fileName, storageFilePath)
	require.NoError(t, err)

	messages, err := messengerService.Check(ctx, mockUsersService, secondUserLogin)
	require.NoError(t, err)
	checkTestFileMessage(t, messages, fileName, storageFilePath, firstUserLogin)

	messegeText := "Hello"
	err = messengerService.Send(ctx, mockUsersService, firstUserLogin, secondUserLogin, messegeText)
	require.NoError(t, err)
	err = messengerService.Send(ctx, mockUsersService, firstUserLogin, secondUserLogin, messegeText)
	require.NoError(t, err)
	err = messengerService.Send(ctx, mockUsersService, firstUserLogin, secondUserLogin, messegeText)
	require.NoError(t, err)
	err = messengerService.Send(ctx, mockUsersService, firstUserLogin, secondUserLogin, messegeText)
	require.NoError(t, err)
	err = messengerService.Send(ctx, mockUsersService, firstUserLogin, secondUserLogin, messegeText)
	require.NoError(t, err)

	messages, err = messengerService.Check(ctx, mockUsersService, secondUserLogin)
	require.NoError(t, err)

	splittedMessages := strings.Split(string(messages), "\n")

	checkTestFileMessage(t, []byte(splittedMessages[0]), fileName, storageFilePath, firstUserLogin)
	checkTestMessage(t, []byte(splittedMessages[1]), messegeText, firstUserLogin)
	checkTestMessage(t, []byte(splittedMessages[2]), messegeText, firstUserLogin)
	checkTestMessage(t, []byte(splittedMessages[3]), messegeText, firstUserLogin)
	checkTestMessage(t, []byte(splittedMessages[4]), messegeText, firstUserLogin)

	err = messengerService.Del(ctx, mockUsersService, secondUserLogin)
	require.NoError(t, err)

	messages, err = messengerService.Check(ctx, mockUsersService, secondUserLogin)
	require.NoError(t, err)
	require.NotContains(t, string(messages), "\n")
	require.Contains(t, string(messages), messegeText)

	err = messengerService.Del(ctx, mockUsersService, secondUserLogin)
	require.NoError(t, err)

	messages, err = messengerService.Check(ctx, mockUsersService, secondUserLogin)
	require.NoError(t, err)
	require.Len(t, string(messages), 0)

	mockUsersService.AssertExpectations(t)
}

func checkTestFileMessage(t *testing.T, messages []byte, fileName, storagePath, senderLogin string) {
	messageStruct := models.Message{}
	json.Unmarshal(messages, &messageStruct)

	require.Equal(t, messageStruct.Text, fmt.Sprintf("p\\%s\\%s", fileName, storagePath))
}

func checkTestMessage(t *testing.T, messages []byte, expectedText, senderLogin string) {
	messageStruct := models.Message{}
	json.Unmarshal(messages, &messageStruct)

	require.Equal(t, messageStruct.Text, expectedText)
	require.Equal(t, messageStruct.SenderLogin, senderLogin)
}
