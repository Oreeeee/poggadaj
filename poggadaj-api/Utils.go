// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"strconv"
)

func GetCookieSafe(c echo.Context, name string) *http.Cookie {
	cookie, _ := c.Cookie(name)
	if cookie != nil {
		return cookie
	}
	return &http.Cookie{}
}

func PasswordFitsRestrictions(password string) bool {
	if len(password) < 8 || len(password) > 64 {
		return false
	}
	return true
}

func GetSeed() uint32 {
	seed64, _ := strconv.ParseUint(os.Getenv("GG_SEED"), 10, 32)
	return uint32(seed64)
}
