package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"bubble/internal/crypto"

	"github.com/labstack/echo/v4"
)

const (
	identificatorForInitAudioDialog           = "a\\"
	identificatorForStopAudioDialog           = "as\\"
	identificatorForSecondStepInitAudioDialog = "ai\\"
)

type RequestCheckMessages struct {
	Gzip     bool
	Messages []byte
}

var audioDialogsMu sync.RWMutex

func (h *Handler) SendMessage(c echo.Context) error {
	var response string
	senderLogin := c.FormValue("senderlogin")
	senderPassword := c.FormValue("senderpassword")
	receiverLogin := c.FormValue("receiverlogin")
	message := c.FormValue("message")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if senderLogin == "" || senderPassword == "" || receiverLogin == "" || message == "" || device == "" || keyForServerDataBase == "" {
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), senderLogin, senderPassword, device)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "GetAuthInfo", "error", err, "ip", c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	senderLogin, err = crypto.StringDecrypt(senderLogin, key)
	if err != nil || senderLogin == "" {
		slog.WarnContext(c.Request().Context(), "Decrypt error", "deviceID", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	receiverLogin, err = crypto.StringDecrypt(receiverLogin, key)
	if err != nil || receiverLogin == "" {
		slog.WarnContext(c.Request().Context(), "Decrypt error", "deviceID", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	message, err = crypto.StringDecrypt(message, key)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "Decrypt error", "deviceID", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if message == identificatorForInitAudioDialog {
		dialogID, senderID, receiverID, err := h.AudioDialogService.CreateAudioDialog(c.Request().Context())
		if err != nil {
			slog.ErrorContext(c.Request().Context(), "Create dialog", "error", err)
			encryptResp, _ := crypto.StringEncrypt([]byte("Create dialog error"), key)
			return c.String(http.StatusInternalServerError, encryptResp)
		}

		message = identificatorForSecondStepInitAudioDialog + dialogID + "\\" + receiverID
		response = dialogID + "\\" + senderID
	} else if strings.HasPrefix(message, identificatorForStopAudioDialog) {
		dialogUUID := strings.Split(message, "\\")[1]
		h.AudioDialogService.DeleteAudioDialog(c.Request().Context(), dialogUUID)
		response = "Audio dialog stop"
	}

	if response != "" {
		responseEncrypted, _ := crypto.StringEncrypt([]byte(response), key)
		return c.String(http.StatusOK, responseEncrypted)
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func (h *Handler) CheckMessage(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || keyForServerDataBase == "" {
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), login, password, device)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "GetAuthInfo", "error", err, "ip", c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	login, err = crypto.StringDecrypt(login, key)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "Decrypt error", "deviceID", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	newMessages, err := h.MessengerService.Check(c.Request().Context(), h.UsersService, login)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "Check Messages", "error", err)
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	messagesEncryptedByte, _ := crypto.StringEncryptByte(newMessages, key)
	request := RequestCheckMessages{
		Messages: messagesEncryptedByte,
	}

	return c.JSON(http.StatusOK, request)
}

func (h *Handler) DelMessages(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || keyForServerDataBase == "" {
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), login, password, device)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "GetAuthInfo", "error", err, "ip", c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	login, err = crypto.StringDecrypt(login, key)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "Decrypt error", "deviceID", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if err := h.MessengerService.Del(c.Request().Context(), h.UsersService, login); err != nil {
		slog.ErrorContext(c.Request().Context(), "Del message", "error", err)
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	return c.NoContent(http.StatusOK)
}
