package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"bubble/internal/crypto"

	"github.com/labstack/echo/v4"
)

type SendMessageRequest struct {
	SenderLogin   string
	ReceiverLogin string
	Message       string
}

type CheckMessagesRequest struct {
	Login string
}

type DelMessagesRequest struct {
	Login string
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

func (h *Handler) SendMessage(c echo.Context, requestBody []byte, key string) error {
	var response string

	var req SendMessageRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.SenderLogin == "" || req.ReceiverLogin == "" || req.Message == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	if req.Message == identificatorForInitAudioDialog {
		dialogID, senderID, receiverID, err := h.AudioDialogService.CreateAudioDialog(c.Request().Context())
		if err != nil {
			c.Logger().Errorf("Create dialog ERR: %s", err)
			resp := "Create dialog error"
			if key != "" {
				resp, _ = crypto.StringEncrypt([]byte(resp), key)
			}
			return c.String(http.StatusInternalServerError, resp)
		}

		req.Message = identificatorForSecondStepInitAudioDialog + dialogID + "\\" + receiverID
		response = dialogID + "\\" + senderID
	} else if strings.HasPrefix(req.Message, identificatorForStopAudioDialog) {
		dialogUUID := strings.Split(req.Message, "\\")[1]
		h.AudioDialogService.DeleteAudioDialog(c.Request().Context(), dialogUUID)
		response = "Audio dialog stop"
	}

	if err := h.MessengerService.Send(c.Request().Context(), h.UsersService, req.SenderLogin, req.ReceiverLogin, req.Message); err != nil {
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	if response != "" {
		if key != "" {
			response, _ = crypto.StringEncrypt([]byte(response), key)
		}
		return c.String(http.StatusOK, response)
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func (h *Handler) CheckMessage(c echo.Context, requestBody []byte, key string) error {
	var req CheckMessagesRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Login == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	newMessages, err := h.MessengerService.Check(c.Request().Context(), h.UsersService, req.Login)
	if err != nil {
		c.Logger().Errorf("Check Messages ERR: %s", err)
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	if key != "" {
		newMessages, _ = crypto.StringEncryptByte(newMessages, key)
	}
	request := RequestCheckMessages{
		Messages: newMessages,
	}

	return c.JSON(http.StatusOK, request)
}

func (h *Handler) DelMessages(c echo.Context, requestBody []byte, key string) error {
	var req DelMessagesRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Login == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	if err := h.MessengerService.Del(c.Request().Context(), h.UsersService, req.Login); err != nil {
		c.Logger().Errorf("Del messages ERR: %s", err)
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	return c.NoContent(http.StatusOK)
}
