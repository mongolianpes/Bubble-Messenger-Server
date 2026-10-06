package files

import (
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveAndGetAndDelFile(t *testing.T) {
	usersServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	ctx := context.Background()
	file := make([]byte, 128)
	rand.Read(file)
	storagePath, err := client.SaveFile(ctx, file)
	require.NoError(t, err)

	resp, err := http.Get("http://localhost:8080/" + storagePath)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, respBytes, file)

	err = client.DelFile(ctx, storagePath)
	require.NoError(t, err)

	resp, err = http.Get("http://localhost:8080/" + storagePath)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBytes, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
}
