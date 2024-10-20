package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func GetCookieSafe(c echo.Context, name string) *http.Cookie {
	cookie, _ := c.Cookie(name)
	if cookie != nil {
		return cookie
	}
	return &http.Cookie{}
}
