package audiodialog

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestMessageAudioDialog struct {
	Time  int64
	Audio []byte
}

func TestCreateAndDeleteDialog(t *testing.T) {
	audioDialogServiceHost = "localhost:8086"
	client, err := NewClient()
	require.NoError(t, err)

	ctx := context.Background()

	dialogID, user1ID, user2ID, err := client.CreateAudioDialog(ctx)
	require.NoError(t, err)

	senderIDAudio := []byte("audio-data")
	body, _ := json.Marshal(map[string]any{
		"dialog_id": dialogID,
		"user_id":   user1ID,
		"message":   senderIDAudio,
	})

	resp, err := http.Post("http://localhost:8080/exchangeaudio", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	body, _ = json.Marshal(map[string]any{
		"dialog_id": dialogID,
		"user_id":   user2ID,
		"message":   []byte("audio-data-other"),
	})

	resp, err = http.Post("http://localhost:8080/exchangeaudio", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	var respMessages map[string][]TestMessageAudioDialog
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(bodyBytes, &respMessages))

	_, ok := respMessages[user2ID]
	require.Equal(t, false, ok)
	_, ok = respMessages[user1ID]
	require.Equal(t, true, ok)
	require.Equal(t, respMessages[user1ID][0].Audio, senderIDAudio)

	err = client.DeleteAudioDialog(ctx, dialogID)
	require.NoError(t, err)

	resp, err = http.Post("http://localhost:8080/exchangeaudio", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusInternalServerError)
}
