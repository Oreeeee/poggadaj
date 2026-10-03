// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/labstack/echo/v5"
)

const AuthCookieSize = 64

var (
	BadPasswordLength     = errors.New("bad password length")
	LogoutWithoutToken    = errors.New("attempted log out without auth token")
	DeletingSessionFailed = errors.New("failed to delete session token")
)

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

// Checks whether the password meets the requirements. Returns an error if the check failed.
// Returns nil if the check passed.
func PasswordMeetsRequirements(password string) error {
	passwordLen := len(password)
	if passwordLen < 8 || passwordLen > 20 {
		return BadPasswordLength
	}
	return nil
}

// Logs the user out from the website
func LogoutUser(c *echo.Context, s *Server) error {
	authToken := GetAuthToken(c)
	if authToken == "" {
		s.logger.Warn("attempted log out without auth token")
		return LogoutWithoutToken
	}

	// First off, delete the session token from the cache server
	err := s.cache.DeleteSession(authToken)
	if err != nil {
		s.logger.Error("failed to delete session token", "err", err)
		return DeletingSessionFailed
	}

	// Next, delete the session cookie from the client
	c.SetCookie(&http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
	})

	return nil
}

// IsInternetExplorer checks if the given user agent string belongs to Internet Explorer
func IsInternetExplorer(userAgent string) bool {
	userAgent = strings.ToLower(userAgent)
	return strings.Contains(userAgent, "msie") || strings.Contains(userAgent, "trident/")
}
