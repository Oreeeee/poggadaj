// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
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
