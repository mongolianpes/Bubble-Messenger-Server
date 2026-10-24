package handlers

import (
	"bubble/internal/crypto"
	"context"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SendFileRequest struct {
	Sender   string
	Receiver string
	FileName string
	File     []byte
}

type DelFileRequest struct {
	Filename string
}

type SetAvatarRequest struct {
	Avatar []byte `json:"avatar"`
	Login  string `json:"login"`
}

type GetAvatarRequest struct {
	LoginForSearch string
}

func (h *Handler) SendFile(c echo.Context, requestBody []byte, key string) error {
	var req SendFileRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Sender == "" || req.Receiver == "" || req.FileName == "" || req.File == nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	fileStoragePath, err := h.FilesService.SaveFile(c.Request().Context(), req.File)
	if err != nil {
		c.Logger().Error("Save file ERR: %s", err)
		resp := "Cant save file"
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	if err := h.MessengerService.SendFile(c.Request().Context(), h.UsersService, req.Sender, req.Receiver, req.FileName, fileStoragePath); err != nil {
		go h.FilesService.DelFile(context.Background(), fileStoragePath)
		c.Logger().Errorf("Send file ERR: %s", err)
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) DelFile(c echo.Context, requestBody []byte, key string) error {
	var req DelFileRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Filename == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	if err := h.FilesService.DelFile(c.Request().Context(), req.Filename); err != nil {
		c.Logger().Error("Del file ERR: %s", err)
		resp := "Cant del file"
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) SetAvatar(c echo.Context, requestBody []byte, key string) error {
	var req SetAvatarRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Login == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	storageAvatarPath, err := h.FilesService.SaveAvatar(c.Request().Context(), req.Avatar)
	if err != nil {
		c.Logger().Error("Save avatar ERR: %s", err)
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	if err := h.UsersService.SetAvatar(c.Request().Context(), req.Login, storageAvatarPath); err != nil {
		c.Logger().Error("Set avatar ERR %s", err)
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusBadRequest, resp)
	}

	c.Logger().Infof("Success save avatar by %s", req.Login)

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetAvatar(c echo.Context, requestBody []byte, key string) error {
	var req GetAvatarRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.LoginForSearch == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	storagePath, err := h.UsersService.GetAvatar(c.Request().Context(), req.LoginForSearch)
	if err != nil {
		c.Logger().Error("Get Avatar ERR: %s", err)
		resp := err.Error()
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusInternalServerError, resp)
	}

	if key != "" {
		storagePath, err = crypto.StringEncrypt([]byte(storagePath), key)
		if err != nil {
			resp := "Decrypted error"
			if key != "" {
				resp, _ = crypto.StringEncrypt([]byte(resp), key)
			}
			return c.String(http.StatusInternalServerError, resp)
		}
	}

	return c.String(http.StatusOK, storagePath)
}
