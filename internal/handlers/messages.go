package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"bubble/internal/crypto"
	"bubble/internal/models"

	"github.com/labstack/echo/v4"
)

type SendMessageRequest struct {
	ReceiverLogin string
	Message       string
}

const ()

type RequestCheckMessages struct {
	Gzip     bool
	Messages []byte
}

var audioDialogsMu sync.RWMutex

func (h *Handler) SendMessage(c echo.Context, requestBody []byte, key string, userLogin string) error {
	var response string

	var req SendMessageRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.ReceiverLogin == "" || req.Message == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	if req.Message == models.IdentificatorInitAudioDialog {
		dialogID, senderID, receiverID, err := h.AudioDialogService.CreateAudioDialog(c.Request().Context())
		if err != nil {
			c.Logger().Errorf("Create dialog ERR: %s", err)
			resp := "Create dialog error"
			if key != "" {
				resp, _ = crypto.StringEncrypt([]byte(resp), key)
			}
			return c.String(http.StatusInternalServerError, resp)
		}

		req.Message = models.IdentificatorSecondStepInitAudioDialog + dialogID + "\\" + receiverID
		response = dialogID + "\\" + senderID
	} else if strings.HasPrefix(req.Message, models.IdentificatorStopAudioDialog) {
		splittedMessage := strings.Split(req.Message, "\\")
		if len(splittedMessage) != 2 {
			req.Message = "\\" + req.Message
		} else {
			dialogUUID := splittedMessage[1]
			h.AudioDialogService.DeleteAudioDialog(c.Request().Context(), dialogUUID)
			response = "Audio dialog stop"
		}
	}

	if strings.HasPrefix(req.Message, models.IdentificatorSecondStepInitAudioDialog) {
		splittedMessage := strings.Split(req.Message, "\\")
		if len(splittedMessage) != 3 {
			req.Message = "\\" + req.Message
		}
	}
	if strings.HasPrefix(req.Message, models.IdentificatorSendFile) {
		splittedMessage := strings.Split(req.Message, "\\")
		if len(splittedMessage) != 3 {
			req.Message = "\\" + req.Message
		}
	}
	if strings.HasPrefix(req.Message, models.IdentificatorInitDialog) {
		splittedMessage := strings.Split(req.Message, "\\")
		if len(splittedMessage) != 2 {
			req.Message = "\\" + req.Message
		}
	}
	if strings.HasPrefix(req.Message, models.IdentificatorAnswerDialog) {
		splittedMessage := strings.Split(req.Message, "\\")
		if len(splittedMessage) != 2 {
			req.Message = "\\" + req.Message
		}
	}

	if err := h.MessengerService.Send(c.Request().Context(), h.UsersService, userLogin, req.ReceiverLogin, req.Message); err != nil {
		c.Logger().Error("Send message ERR: %s", err)
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

func (h *Handler) CheckMessage(c echo.Context, key string, userLogin string) error {
	newMessages, err := h.MessengerService.Check(c.Request().Context(), h.UsersService, userLogin)
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

func (h *Handler) DelMessages(c echo.Context, key string, userLogin string) error {
	if err := h.MessengerService.Del(c.Request().Context(), h.UsersService, userLogin); err != nil {
		c.Logger().Errorf("Del messages ERR: %s", err)
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	return c.NoContent(http.StatusOK)
}
