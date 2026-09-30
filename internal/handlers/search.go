package handlers

import (
	"bubble/internal/crypto"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) SearchUser(c echo.Context) error {
	device := c.FormValue("device")
	loginForSearch := c.FormValue("loginforsearch")
	keyForServerDataBase := c.FormValue("forserver")
	login := c.FormValue("login")
	password := c.FormValue("password")
	if device == "" || loginForSearch == "" || keyForServerDataBase == "" || login == "" || password == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), login, password, device)
	if err != nil {
		c.Logger().Errorf("GetAuthInfo IP %s ERR: %s", err, c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	loginForSearch, err = crypto.StringDecrypt(loginForSearch, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	findUsers, err := h.UsersService.Search(c.Request().Context(), loginForSearch)
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
