package handlers

import (
	"net/http"

	"bubble/internal/crypto"

	"github.com/labstack/echo/v4"
)

type KeyExchangeResponse struct {
	ServerPublicKey string `json:"public_key"`
}

type KeyExchangeRequest struct {
	ID              string `json:"id"`
	ClientPublicKey string `json:"public_key"`
	IsRegistring    bool   `json:"is_registring"`
}

func (h *Handler) TLS(c echo.Context) error {
	var req KeyExchangeRequest
	if err := c.Bind(&req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	serverPublicKey, err := h.UsersService.TLS(c.Request().Context(), req.IsRegistring, req.ClientPublicKey, req.ID)
	if err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	response := KeyExchangeResponse{
		ServerPublicKey: serverPublicKey,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) Reg(c echo.Context) error {
	login := c.FormValue("login")
	name := c.FormValue("name")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if login == "" || name == "" || password == "" || device == "" || keyForServerDataBase == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, err := h.UsersService.Register(c.Request().Context(), login, name, password, device)
	if err != nil {
		resp := ""
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(err.Error()), key)
		} else {
			resp = err.Error()
		}

		c.Logger().Warnf("Register error. Login %s, deviceID %s, IP %s, ERR: %s", login, device, c.RealIP(), err)

		return c.String(http.StatusInternalServerError, resp)
	}

	c.Logger().Infof("Success registraion %s. DeviceID %s, IP %s", login, device, c.RealIP())
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Auth(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if login == "" || password == "" || device == "" || keyForServerDataBase == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	userName, key, err := h.UsersService.Auth(c.Request().Context(), login, password, device)
	if err != nil {
		resp := ""
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(err.Error()), key)
		} else {
			resp = err.Error()
		}

		c.Logger().Warnf("Auth error. Login %s, deviceID %s, IP %s, ERR: %s", login, device, c.RealIP(), err)

		return c.String(http.StatusInternalServerError, resp)
	}

	c.Logger().Infof("Success auth %s. DeviceID %s, IP %s", login, device, c.RealIP())

	resp, _ := crypto.StringEncrypt([]byte(userName), key)
	return c.String(http.StatusOK, resp)
}
