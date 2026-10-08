package users

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

func StringEncryptTest(stringToEncrypt []byte, keyString string) (encryptedString string, errEncrypt error) {
	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	// plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, stringToEncrypt, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func ComputeTestSharedSecret(clientPrivateKey *ecdh.PrivateKey, serverPublicKeyB64 string) (string, error) {
	// 1. Декодируем base64.RawURLEncoding → []byte
	serverPubKeyBytes, err := base64.RawURLEncoding.DecodeString(serverPublicKeyB64)
	if err != nil {
		return "", fmt.Errorf("invalid base64 public key: %w", err)
	}

	// 2. Создаём объект публичного ключа
	serverPublicKey, err := ecdh.P256().NewPublicKey(serverPubKeyBytes)
	if err != nil {
		return "", fmt.Errorf("invalid public key: %w", err)
	}

	// 3. Вычисляем shared secret
	sharedSecret, err := clientPrivateKey.ECDH(serverPublicKey)
	if err != nil {
		return "", fmt.Errorf("ECDH failed: %w", err)
	}

	// 4. Возвращаем в hex (как делает сервер)
	return hex.EncodeToString(sharedSecret), nil
}

func TLSTestHandshake(ctx context.Context, client *Client, isRegister bool, device string) (string, error) {
	clientPrivateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	clientPubKeyBytes := clientPrivateKey.PublicKey().Bytes()
	clientPubKeyB64 := base64.RawURLEncoding.EncodeToString(clientPubKeyBytes)

	serverPublicKey, err := client.TLS(ctx, isRegister, clientPubKeyB64, device)
	if err != nil {
		return "", err
	}
	if serverPublicKey == "" {
		return "", errors.New("пустой публичный ключ сервера")
	}

	sharedSecret, err := ComputeTestSharedSecret(clientPrivateKey, serverPublicKey)
	if err != nil {
		return "", err
	}

	return sharedSecret, nil
}

func getTestDSN() string {
	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "5432")
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

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", getTestDSN())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, db.PingContext(ctx))

	_, err = db.ExecContext(ctx, `TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	return db
}

func TestTLSAndRegister(t *testing.T) {
	usersServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	db := setupTestDB(t)

	ctx := context.Background()

	t.Run("success register", func(t *testing.T) {
		login := "user1"
		name := "user1"
		password := "password123"
		device := "device1"

		sharedSecret, err := TLSTestHandshake(ctx, client, true, device)
		require.NoError(t, err)
		loginEncrypt, err := StringEncryptTest([]byte(login), sharedSecret)
		require.NoError(t, err)
		nameEncrypt, err := StringEncryptTest([]byte(name), sharedSecret)
		require.NoError(t, err)
		passwordEncrypt, err := StringEncryptTest([]byte(password), sharedSecret)
		require.NoError(t, err)

		key, err := client.Register(ctx, loginEncrypt, nameEncrypt, passwordEncrypt, device)
		require.NoError(t, err)
		require.NotEmpty(t, key)

		dbLogin := ""
		err = db.QueryRow("SELECT login FROM users WHERE login = $1", login).Scan(&dbLogin)
		require.NoError(t, err)
		require.Equal(t, dbLogin, login)
	})

	t.Run("register without encode data", func(t *testing.T) {
		login := "user1"
		name := "user1"
		password := "password123"
		device := "device1"

		_, err = client.Register(ctx, login, name, password, device)
		require.Error(t, err)
	})

	t.Run("register with already exists login", func(t *testing.T) {
		login := "user2"
		name := "user2"
		password := "password123"
		device := "device2"

		sharedSecret, err := TLSTestHandshake(ctx, client, true, device)
		require.NoError(t, err)
		loginEncrypt, err := StringEncryptTest([]byte(login), sharedSecret)
		require.NoError(t, err)
		nameEncrypt, err := StringEncryptTest([]byte(name), sharedSecret)
		require.NoError(t, err)
		passwordEncrypt, err := StringEncryptTest([]byte(password), sharedSecret)
		require.NoError(t, err)

		_, err = db.Exec("INSERT INTO users (login, name, password) VALUES ($1, $2, $3)", login, name, password)
		require.NoError(t, err)

		_, err = client.Register(ctx, loginEncrypt, nameEncrypt, passwordEncrypt, device)
		require.Error(t, err)
	})
}

func TestTLSAndAuth(t *testing.T) {
	usersServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	ctx := context.Background()

	login := "user1"
	name := "user1"
	password := "password123"
	device := "device1"

	sharedSecret, err := TLSTestHandshake(ctx, client, false, device)
	require.NoError(t, err)
	loginEncrypt, err := StringEncryptTest([]byte(login), sharedSecret)
	require.NoError(t, err)
	passwordEncrypt, err := StringEncryptTest([]byte(password), sharedSecret)
	require.NoError(t, err)

	nameAuth, key, err := client.Auth(ctx, loginEncrypt, passwordEncrypt, device)
	require.NoError(t, err)
	require.Equal(t, nameAuth, name)
	require.NotEmpty(t, key)
}

func TestTLSAndGetAuthInfo(t *testing.T) {
	usersServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	ctx := context.Background()

	login := "user1"
	password := "password123"
	device := "device1"

	sharedSecret, err := TLSTestHandshake(ctx, client, false, device)
	require.NoError(t, err)
	loginEncrypt, err := StringEncryptTest([]byte(login+"1"), sharedSecret)
	require.NoError(t, err)
	passwordEncrypt, err := StringEncryptTest([]byte(password), sharedSecret)
	require.NoError(t, err)

	_, _, err = client.Auth(ctx, loginEncrypt, passwordEncrypt, device)
	require.Error(t, err)

	key, userID, err := client.GetAuthInfo(ctx, login, password, device)
	require.NoError(t, err)
	require.NotEmpty(t, key)
	require.NotEmpty(t, userID)
}
