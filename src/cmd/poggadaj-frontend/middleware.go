// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"github.com/labstack/echo/v5"
	"github.com/strukturag/goacceptlanguageparser"
)

// LanguageMiddleware selects a best-fit language for the user, and sets it to the "poggadaj-lang" key of the Echo context
func LanguageMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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
