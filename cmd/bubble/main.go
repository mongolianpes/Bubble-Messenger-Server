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
		mainService := echo.New()
		mainService.Use(IPBlockMiddleware(hand.Blocker))
		mainService.Use(middleware.Recover())
		mainService.Use(middleware.ContextTimeout(timeToProcessRequest))
		mainService.HideBanner = true
		mainService.POST("/exchangekey", hand.TLS)
		mainService.POST("/reg", hand.Reg)
		mainService.POST("/auth", hand.Auth)
		mainService.POST("/searchuser", hand.SearchUser)
		mainService.POST("/sendmessage", hand.SendMessage)
		mainService.POST("/checkmessage", hand.CheckMessage)

		if err := mainService.Start(":23099"); err != nil {
			mainService.Logger.Error("Start main Service: %s", err)
		}
	}()

	go func() {
		mediumSizeDataService := echo.New()
		mediumSizeDataService.Use(IPBlockMiddleware(hand.Blocker))
		mediumSizeDataService.Use(middleware.Recover())
		mediumSizeDataService.Use(middleware.ContextTimeout(timeToProcessRequest))
		mediumSizeDataService.HideBanner = true
		mediumSizeDataService.POST("/delmessages", hand.DelMessages)
		mediumSizeDataService.POST("/setavatar", hand.SetAvatar)
		mediumSizeDataService.POST("/getavatar", hand.GetAvatar)
		mediumSizeDataService.POST("/sendfile", hand.SendFile)

		if err := mediumSizeDataService.Start(":23098"); err != nil {
			mediumSizeDataService.Logger.Errorf("Start medium size Data Service: %s", err)
		}
	}()
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
