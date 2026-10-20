package handlers

import (
	"net/http"
	"strings"
	"sync"

	"bubble/internal/crypto"

	"github.com/labstack/echo/v4"
)

type SendMessageRequest struct {
	SenderLogin    string
	SenderPassword string
	ReceiverLogin  string
	Device         string
	Message        string
}

type CheckMessagesRequest struct {
	Login    string
	Password string
	Device   string
}

type DelMessagesRequest struct {
	Login    string
	Password string
	Device   string
}

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

	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.SenderLogin == "" || req.SenderPassword == "" || req.ReceiverLogin == "" || req.Message == "" || req.Device == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), req.SenderLogin, req.SenderPassword, req.Device)
	if err != nil {
		c.Logger().Errorf("GetAuthInfo IP %s ERR: %s", err, c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	req.SenderLogin, err = crypto.StringDecrypt(req.SenderLogin, key)
	if err != nil || req.SenderLogin == "" {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	req.ReceiverLogin, err = crypto.StringDecrypt(req.ReceiverLogin, key)
	if err != nil || req.ReceiverLogin == "" {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	req.Message, err = crypto.StringDecrypt(req.Message, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if req.Message == identificatorForInitAudioDialog {
		dialogID, senderID, receiverID, err := h.AudioDialogService.CreateAudioDialog(c.Request().Context())
		if err != nil {
			c.Logger().Errorf("Create dialog ERR: %s", err)
			encryptResp, _ := crypto.StringEncrypt([]byte("Create dialog error"), key)
			return c.String(http.StatusInternalServerError, encryptResp)
		}

		req.Message = identificatorForSecondStepInitAudioDialog + dialogID + "\\" + receiverID
		response = dialogID + "\\" + senderID
	} else if strings.HasPrefix(req.Message, identificatorForStopAudioDialog) {
		dialogUUID := strings.Split(req.Message, "\\")[1]
		h.AudioDialogService.DeleteAudioDialog(c.Request().Context(), dialogUUID)
		response = "Audio dialog stop"
	}

	if err = h.MessengerService.Send(c.Request().Context(), h.UsersService, req.SenderLogin, req.ReceiverLogin, req.Message); err != nil {
		responseEncrypted, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, responseEncrypted)
	}

	if response != "" {
		responseEncrypted, _ := crypto.StringEncrypt([]byte(response), key)
		return c.String(http.StatusOK, responseEncrypted)
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func (h *Handler) CheckMessage(c echo.Context) error {
	var req CheckMessagesRequest
	if err := c.Bind(&req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Device == "" || req.Login == "" || req.Password == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), req.Login, req.Password, req.Device)
	if err != nil {
		c.Logger().Errorf("GetAuthInfo IP %s ERR: %s", err, c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	newMessages, err := h.MessengerService.Check(c.Request().Context(), h.UsersService, req.Login)
	if err != nil {
		c.Logger().Errorf("Check Messages ERR: %s", err)
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
	var req DelMessagesRequest
	if err := c.Bind(&req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Device == "" || req.Login == "" || req.Password == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), req.Login, req.Password, req.Device)
	if err != nil {
		c.Logger().Errorf("GetAuthInfo IP %s ERR: %s", err, c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if err := h.MessengerService.Del(c.Request().Context(), h.UsersService, req.Login); err != nil {
		c.Logger().Errorf("Del messages ERR: %s", err)
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	return c.NoContent(http.StatusOK)
}
