package cookie

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	UserLogin = "login"
	UserName  = "name"
	SessionID = "session_id"
)

func Set(c echo.Context, name, value string, useHTTPOnly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: useHTTPOnly,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}
