// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"net/mail"

	"github.com/labstack/echo/v5"
)

const AuthCookieSize = 64

// GenerateAuthToken generates a random cryptographically secure auth token with AuthCookieSize bytes of entropy
func GenerateAuthToken() (string, error) {
	authCookie := make([]byte, AuthCookieSize)
	if _, err := crand.Read(authCookie); err != nil {
		return "", errors.New("failed to generate auth token")
	}
	return base64.RawURLEncoding.EncodeToString(authCookie), nil
}

// VerifyEmail checks if the email format is correct
func VerifyEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Gets a query param by name and parses it into a bool if name is "1"
func QueryParamBool(c *echo.Context, name string) bool {
	value := c.QueryParamOr(name, "0")
	if value == "1" {
		return true
	}
	return false
}
