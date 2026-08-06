// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func GetCookieSafe(c *echo.Context, name string) *http.Cookie {
	cookie, _ := c.Cookie(name)
	if cookie != nil {
		return cookie
	}
	return &http.Cookie{}
}
