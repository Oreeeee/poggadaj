// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/strukturag/goacceptlanguageparser"
)

type MiddlewareController struct {
	server *Server
}

// LanguageMiddleware selects a best-fit language for the user, and sets it to the "poggadaj-lang" key of the Echo context
func (mc *MiddlewareController) LanguageMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		acceptLanguageHdr := c.Request().Header.Get("Accept-Language")
		browserLanguages := goacceptlanguageparser.ParseAcceptLanguage(acceptLanguageHdr, []string{"en", "pl"}) // TODO: Properly retrieve the list of supported languages

		if len(browserLanguages) == 0 { // Couldn't find a best fit language
			c.Set("poggadaj-lang", "en")
			return next(c)
		}

		c.Set("poggadaj-lang", browserLanguages[0])
		return next(c)
	}
}

// HasAuthMiddleware checks whether the user has a valid login session and sets a flag in the echo context.
// It does not enforce security for protected resources! It only checks if the session is valid.
func (mc *MiddlewareController) HasAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		c.Set("poggadaj-has-auth", false)
		c.Set("poggadaj-uin", 0)

		authCookie, err := c.Cookie("token")
		if err != nil {
			mc.server.logger.Warn("couldn't get the session cookie", "err", err)
			return next(c)
		}

		uinCookie, err := c.Cookie("uin")
		if err != nil {
			mc.server.logger.Warn("couldn't get the uin cookie", "err", err)
			return next(c)
		}

		uin64, err := strconv.ParseUint(uinCookie.Value, 10, 32)
		if err != nil {
			mc.server.logger.Error("couldn't read uin from cookie", "err", err)
			return next(c)
		}
		uin := uint(uin64)

		authToken := authCookie.Value

		tokenOwnerUin, err := mc.server.cache.GetSessionTokenOwner(authToken)
		if err != nil {
			mc.server.logger.Error("couldn't verify the token", "err", err)
			return next(c)
		}

		// Verify the token ownership
		if uin != tokenOwnerUin {
			mc.server.logger.Warn("mismatch between uin cookie and token owner", "uin", uinCookie)
			return next(c)
		}

		c.Set("poggadaj-auth-token", authToken)
		c.Set("poggadaj-has-auth", true)
		c.Set("poggadaj-uin", tokenOwnerUin)

		return next(c)
	}
}

func NewMiddlewareController(server *Server) (*MiddlewareController, error) {
	return &MiddlewareController{server: server}, nil
}
