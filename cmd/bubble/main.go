package main

import (
	"fmt"
	"net/http"
	"time"

	"bubble/internal/handlers"
	"bubble/internal/ipblocker"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const timeToProcessRequest = 10 * time.Second

func main() {
	fmt.Printf(`
   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v%s
Start on ports: 23099, 23098, 23097
`, echo.Version)

	hand, err := handlers.NewHand()
	if err != nil {
		panic(err)
	}

	go func() {
		e := echo.New()
		e.Use(IPBlockMiddleware(hand.Blocker))
		e.Use(middleware.Recover())
		e.Use(middleware.ContextTimeout(timeToProcessRequest))
		e.HideBanner = true
		e.POST("/reg", hand.Handlers.Reg)
		e.POST("/auth", hand.Handlers.Auth)
		e.POST("/searchuser", hand.RouterSearchUser)
		e.POST("/sendmessage", hand.RouterSendMessage)
		e.POST("/checkmessage", hand.RouterCheckMessages)
		e.POST("/delmessages", hand.RouterDelMessages)
		e.POST("/setavatar", hand.RouterSetAvatar)
		e.POST("/getavatar", hand.RouterGetAvatar)
		e.POST("/sendfile", hand.RouterSendFile)
		e.POST("/delfile", hand.RouterDelFile)

		if err := e.Start(":443"); err != nil {
			e.Logger.Error("Start https Service: %s", err)
		}
	}()

	go func() {
		e := echo.New()
		e.Use(IPBlockMiddleware(hand.Blocker))
		e.Use(middleware.Recover())
		e.Use(middleware.ContextTimeout(timeToProcessRequest))
		e.HideBanner = true
		e.POST("/*", hand.Router)
		e.POST("/exchangekey", hand.Handlers.TLS)
		e.POST("/reg", hand.Handlers.Reg)
		e.POST("/auth", hand.Handlers.Auth)
		// e.POST("/searchuser", hand.SearchUser)
		// e.POST("/sendmessage", hand.Handlers.SendMessage)
		// e.POST("/checkmessage", hand.Handlers.CheckMessage)
		// e.POST("/delmessages", hand.Handlers.DelMessages)

		if err := e.Start(":23099"); err != nil {
			e.Logger.Error("Start main Service: %s", err)
		}
	}()

	go func() {
		e := echo.New()
		e.Use(IPBlockMiddleware(hand.Blocker))
		e.Use(middleware.Recover())
		e.Use(middleware.ContextTimeout(timeToProcessRequest))
		e.HideBanner = true
		e.POST("/*", hand.Router)
		// e.POST("/setavatar", hand.Handlers.SetAvatar)
		// e.POST("/getavatar", hand.Handlers.GetAvatar)
		// e.POST("/sendfile", hand.Handlers.SendFile)
		// e.POST("/delfile", hand.Handlers.DelFile)

		if err := e.Start(":23098"); err != nil {
			e.Logger.Errorf("Start medium size Data Service: %s", err)
		}
	}()

	select {}
}

func IPBlockMiddleware(blocker ipblocker.Blocker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if blocker.IsBlocked(ip) {
				return c.String(http.StatusForbidden, "Your IP is temporarily blocked")
			}
			return next(c)
		}
	}
}
