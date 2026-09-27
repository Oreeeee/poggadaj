// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import "github.com/labstack/echo/v5"

// Retrieves the user's language from the request context. Returns "en" if empty.
func GetLang(c *echo.Context) string {
	val := c.Get("poggadaj-lang")
	if val == nil {
		return "en"
	}
	return val.(string)
}

// Retrieves user's authentication status from the request context. Returns false if empty.
func GetHasAuth(c *echo.Context) bool {
	if hasAuthRaw := c.Get("poggadaj-has-auth"); hasAuthRaw != nil {
		return hasAuthRaw.(bool)
	}
	return false
}

// Retrieves user's uin from the request context. Returns 0 if empty.
func GetUin(c *echo.Context) uint {
	if uinRaw := c.Get("poggadaj-uin"); uinRaw != nil {
		return uinRaw.(uint)
	}
	return 0
}

// Retrieves user's auth token from the request context. Returns an empty string if empty.
func GetAuthToken(c *echo.Context) string {
	val := c.Get("poggadaj-auth-token")
	if val == nil {
		return ""
	}
	return val.(string)
}

func SetLang(c *echo.Context, val string) {
	c.Set("poggadaj-lang", val)
}

func SetHasAuth(c *echo.Context, val bool) {
	c.Set("poggadaj-has-auth", val)
}

func SetUin(c *echo.Context, uin uint) {
	c.Set("poggadaj-uin", uin)
}

func SetAuthToken(c *echo.Context, val string) {
	c.Set("poggadaj-auth-token", val)
}
