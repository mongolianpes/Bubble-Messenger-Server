package handlers

import (
	"bubble/internal/cookie"
	"bubble/internal/crypto"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const maxRequestLatency = 30 * time.Second

type TLSData struct {
	Device      string
	Login       string
	Password    string
	SendTime    string
	Nonce       string
	RequestBody []byte
}

func (rh *TLSHandler) Router(c echo.Context) error {
	var req TLSData
	if err := c.Bind(&req); err != nil {
		rh.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Device == "" || req.Login == "" || req.Password == "" {
		rh.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := rh.Auth.GetAuthInfo(c.Request().Context(), req.Login, req.Password, req.Device)
	if err != nil {
		c.Logger().Errorf("GetAuthInfo IP %s ERR: %s", err, c.RealIP())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	req.SendTime, err = crypto.StringDecrypt(req.SendTime, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	if isOld, err := checkRequestLatencyOld(req.SendTime); isOld || err != nil {
		if isOld {
			c.Logger().Warnf("Send time old. DeviceID %s, SendTime %s", req.Device, req.SendTime)
			encryptResp, _ := crypto.StringEncrypt([]byte("Old request"), key)
			return c.String(http.StatusBadRequest, encryptResp)
		}

		if err != nil {
			c.Logger().Warnf("Cant parse send time. DeviceID %s, send time %s", req.Device, req.SendTime)
			encryptResp, _ := crypto.StringEncrypt([]byte("Can not parse send time"), key)
			return c.String(http.StatusBadRequest, encryptResp)
		}
	}

	req.Nonce, err = crypto.StringDecrypt(req.Nonce, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if nonce, ok := rh.UsersUsedRequestsNonces[req.Login]; ok {
		if nonce == req.Nonce {
			c.Logger().Warnf("Nonce is used twice. DeviceID %s, SendTime %s, Nonce %s", req.Device, req.SendTime, req.Nonce)
			encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
			return c.String(http.StatusBadRequest, encryptResp)
		}
	}

	rh.UsersUsedRequestsNonces[req.Login] = req.Nonce

	req.RequestBody, err = crypto.StringDecryptByte(req.RequestBody, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	URLPath := strings.TrimPrefix(c.Request().URL.Path, "/")
	URLPath, err = crypto.StringDecrypt(URLPath, key)
	if err != nil {
		c.Logger().Warnf("Decrypt error. DeviceID %s", req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	switch URLPath {
	case "searchuser":
		return rh.Handlers.SearchUser(c, req.RequestBody, key)
	case "sendmessage":
		return rh.Handlers.SendMessage(c, req.RequestBody, key, req.Login)
	case "checkmessage":
		return rh.Handlers.CheckMessage(c, key, req.Login)
	case "delmessages":
		return rh.Handlers.DelMessages(c, key, req.Login)
	case "setavatar":
		return rh.Handlers.SetAvatar(c, req.RequestBody, key, req.Login)
	case "getavatar":
		return rh.Handlers.GetAvatar(c, req.RequestBody, key)
	case "sendfile":
		return rh.Handlers.SendFile(c, req.RequestBody, key, req.Login)
	case "delfile":
		return rh.Handlers.DelFile(c, req.RequestBody, key)
	default:
		return c.NoContent(http.StatusNotFound)
	}
}

func checkRequestLatencyOld(sendTime string) (latencyOld bool, err error) {
	latencyOld = true

	var reqSendTime time.Time
	reqSendTime, err = time.Parse(time.RFC3339, sendTime)
	if err != nil {
		return
	}

	if !reqSendTime.Add(maxRequestLatency).Before(time.Now()) {
		latencyOld = false
	}
	return
}

func (rh *TLSHandler) RouterSendFile(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	login, err := c.Cookie(cookie.UserLogin)
	if err != nil {
		return c.String(http.StatusBadRequest, "Cookie mycookie is missing")
	}

	return rh.Handlers.SendFile(c, body, "", login.Value)
}

func (rh *TLSHandler) RouterDelFile(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	return rh.Handlers.DelFile(c, body, "")
}

func (rh *TLSHandler) RouterSetAvatar(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	login, err := c.Cookie(cookie.UserLogin)
	if err != nil {
		return c.String(http.StatusBadRequest, "Cookie mycookie is missing")
	}

	return rh.Handlers.SetAvatar(c, body, "", login.Value)
}

func (rh *TLSHandler) RouterGetAvatar(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	return rh.Handlers.GetAvatar(c, body, "")
}

func (rh *TLSHandler) RouterSearchUser(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	return rh.Handlers.SearchUser(c, body, "")
}

func (rh *TLSHandler) RouterSendMessage(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	login, err := c.Cookie(cookie.UserLogin)
	if err != nil {
		return c.String(http.StatusBadRequest, "Cookie mycookie is missing")
	}

	return rh.Handlers.SendMessage(c, body, "", login.Value)
}

func (rh *TLSHandler) RouterCheckMessages(c echo.Context) error {
	login, err := c.Cookie(cookie.UserLogin)
	if err != nil {
		return c.String(http.StatusBadRequest, "Cookie mycookie is missing")
	}

	return rh.Handlers.CheckMessage(c, "", login.Value)
}

func (rh *TLSHandler) RouterDelMessages(c echo.Context) error {
	login, err := c.Cookie(cookie.UserLogin)
	if err != nil {
		return c.String(http.StatusBadRequest, "Cookie mycookie is missing")
	}

	return rh.Handlers.DelMessages(c, "", login.Value)
}
