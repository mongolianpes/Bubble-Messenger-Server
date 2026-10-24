package handlers

import (
	"bubble/internal/audiodialog"
	"bubble/internal/crypto"
	"bubble/internal/files"
	"bubble/internal/ipblocker"
	"bubble/internal/messenger"
	"bubble/internal/users"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const maxRequestLatency = 30 * time.Second

type Handler struct {
	UsersService       users.UsersService
	MessengerService   messenger.MessengerService
	FilesService       files.FilesService
	AudioDialogService audiodialog.AudioDialogService
	Blocker            ipblocker.Blocker
}

type TLSHandler struct {
	Handlers                *Handler
	Auth                    users.Auth
	Blocker                 ipblocker.Blocker
	UsersUsedRequestsNonces map[string]string
}

type TLSData struct {
	Device      string
	Login       string
	Password    string
	SendTime    string
	Nonce       string
	RequestBody []byte
}

var invalidJSONerror = errors.New("Invalid JSON")

func NewHand() (*TLSHandler, error) {
	usersService, err := users.NewClient()
	if err != nil {
		return nil, err
	}

	messengerService, err := messenger.NewClient()
	if err != nil {
		return nil, err
	}

	filesService, err := files.NewClient()
	if err != nil {
		return nil, err
	}

	audioDialogService, err := audiodialog.NewClient()
	if err != nil {
		return nil, err
	}

	blocker := ipblocker.New(10, 30*time.Minute)

	router := &TLSHandler{
		Handlers: &Handler{
			UsersService:       usersService,
			MessengerService:   messengerService,
			FilesService:       filesService,
			AudioDialogService: audioDialogService,
		},
		Blocker:                 blocker,
		Auth:                    usersService,
		UsersUsedRequestsNonces: make(map[string]string),
	}

	return router, nil
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
		return rh.Handlers.SendMessage(c, req.RequestBody, key)
	case "checkmessage":
		return rh.Handlers.CheckMessage(c, req.RequestBody, key)
	case "delmessages":
		return rh.Handlers.DelMessages(c, req.RequestBody, key)
	case "setavatar":
		return rh.Handlers.SetAvatar(c, req.RequestBody, key)
	case "getavatar":
		return rh.Handlers.GetAvatar(c, req.RequestBody, key)
	case "sendfile":
		return rh.Handlers.SendFile(c, req.RequestBody, key)
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

	return rh.Handlers.SendFile(c, body, "")
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

	return rh.Handlers.SetAvatar(c, body, "")
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

	return rh.Handlers.SendMessage(c, body, "")
}

func (rh *TLSHandler) RouterCheckMessages(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	return rh.Handlers.CheckMessage(c, body, "")
}

func (rh *TLSHandler) RouterDelMessages(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	return rh.Handlers.DelMessages(c, body, "")
}
