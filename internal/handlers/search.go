package handlers

import (
	"bubble/internal/crypto"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SearchUserRequest struct {
	LoginForSearch string
}

func (h *Handler) SearchUser(c echo.Context, requestBody []byte, key string) error {
	var req SearchUserRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.LoginForSearch == "" {
		h.Blocker.RegisterStrike(c.RealIP())
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	findUsers, err := h.UsersService.Search(c.Request().Context(), req.LoginForSearch)
	if err != nil {
		resp := "Can not find users"
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusBadRequest, resp)
	}

	b, err := json.Marshal(findUsers)
	if err != nil {
		c.Logger().Errorf("Marshal JSON ERR: %s", err)
		resp := "Can not marshal find users"
		if key != "" {
			resp, _ = crypto.StringEncrypt([]byte(resp), key)
		}
		return c.String(http.StatusBadRequest, resp)
	}

	if key != "" {
		b, _ = crypto.StringEncryptByte(b, key)
	}

	return c.Blob(http.StatusOK, "application/octet-stream", b)
}
