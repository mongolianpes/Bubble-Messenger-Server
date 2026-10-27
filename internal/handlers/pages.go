package handlers

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

const (
	homeHTMLPagePath  = "pages/home.html"
	authHTMLPagePath  = "pages/auth.html"
	chatsHTMLPagePath = "pages/chats.html"
)

func (h *PagesHandler) HomePage(c echo.Context) error {
	return c.HTML(http.StatusOK, h.homeHTMLPage)
}

func (h *PagesHandler) AuthPage(c echo.Context) error {
	return c.HTML(http.StatusOK, h.authHTMLPage)
}

func (h *PagesHandler) ChatsPage(c echo.Context) error {
	return c.HTML(http.StatusOK, h.chatsHTMLPage)
}

func (h *PagesHandler) readHTMLPages() error {
	homeHTML, err := os.ReadFile(homeHTMLPagePath)
	if err != nil {
		return err
	}
	h.homeHTMLPage = string(homeHTML)

	authHTML, err := os.ReadFile(authHTMLPagePath)
	if err != nil {
		return err
	}
	h.authHTMLPage = string(authHTML)

	chatsHTML, err := os.ReadFile(chatsHTMLPagePath)
	if err != nil {
		return err
	}
	h.chatsHTMLPage = string(chatsHTML)

	return nil
}
