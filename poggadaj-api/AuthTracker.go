package main

import (
	"github.com/labstack/echo/v4"
	"math/rand"
	"time"
)

const AuthCookieSize = 64
const AuthCookieChars = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM0123456789"
const AuthCookieLifetime = 2 * time.Hour

type AuthorizedSession struct {
	User       string
	AuthCookie string
	Expires    time.Time
}

func GenerateAuthorizedSession(user string) AuthorizedSession {
	authCookie := make([]byte, AuthCookieSize)
	for i := range authCookie {
		authCookie[i] = AuthCookieChars[rand.Intn(len(AuthCookieChars))]
	}
	expires := time.Now().Add(AuthCookieLifetime)
	return AuthorizedSession{
		User:       user,
		AuthCookie: string(authCookie),
		Expires:    expires,
	}
}

func ValidateSession(c echo.Context) bool {
	nameCookie := GetCookieSafe(c, "Username")
	authCookie := GetCookieSafe(c, "Auth")

	for i := range Sessions {
		session := Sessions[i]
		userMatches := session.User == nameCookie.Value
		cookieMatches := session.AuthCookie == authCookie.Value
		if userMatches && cookieMatches && authCookie.Expires.Unix() < time.Now().Unix() {
			return true
		}
	}
	return false
}
