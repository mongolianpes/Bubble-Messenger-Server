package handlers

import (
	"bubble/internal/crypto"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SearchUserRequest struct {
	Login          string
	Password       string
	Device         string
	LoginForSearch string
}

func (h *Handler) SearchUser(c echo.Context) error {
	var req SearchUserRequest
	if err := c.Bind(&req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Device == "" || req.LoginForSearch == "" || req.Login == "" || req.Password == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), req.Login, req.Password, req.Device)
	if err != nil {
		c.Logger().Errorf("GetAuthInfo IP %s ERR: %s", err, c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	req.LoginForSearch, err = crypto.StringDecrypt(req.LoginForSearch, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	findUsers, err := h.UsersService.Search(c.Request().Context(), req.LoginForSearch)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Can not find users"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	b, err := json.Marshal(findUsers)
	if err != nil {
		c.Logger().Errorf("Marshal JSON ERR: %s", err)
		encryptResp, _ := crypto.StringEncrypt([]byte("Can not marshal find users"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	b, _ = crypto.StringEncryptByte(b, key)

	return c.Blob(http.StatusOK, "application/octet-stream", b)
}
