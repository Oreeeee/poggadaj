// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package server

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"codeberg.org/or3e/poggadaj/internal/security/argon2"
	"codeberg.org/or3e/poggadaj/internal/security/gg"
	"codeberg.org/or3e/poggadaj/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v5"
)

func (s *Server) handleLogin(c *echo.Context) error {
	data := map[string]any{}

	data["showLoginFail"] = QueryParamBool(c, "fail")
	data["showLogoutSuccessful"] = QueryParamBool(c, "loggedOut")
	data["showRegisterSuccess"] = QueryParamBool(c, "registerSuccess")
	data["passwordChanged"] = QueryParamBool(c, "passwordChanged")

	return c.Render(http.StatusOK, "login.jet", data)
}

func (s *Server) handleLoginAction(c *echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Get the actual password hash from the database
	uin, passwordHash, err := s.db.Users.GetUserPasswordHashWithUin(username)

	if errors.Is(err, pgx.ErrNoRows) {
		// The user doesn't even exist
		return c.Redirect(http.StatusSeeOther, "/login?fail=1")
	}

	if err != nil {
		s.logger.Error("failed to get password from database", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// And then verify if it matches with the one the user provided
	match, err := argon2.ComparePasswords(password, passwordHash)
	if err != nil {
		s.logger.Error("failed to compare passwords", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if !match {
		return c.Redirect(http.StatusSeeOther, "/login?fail=1")
	}

	// Generate, cache, and send a session token
	var ttl time.Duration = time.Hour * 7 * 24
	token, err := GenerateAuthToken()
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	err = s.cache.CreateFrontendSession(token, uin, ttl)
	if err != nil {
		s.logger.Error("failed to save the session token", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	c.SetCookie(&http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(ttl),
	})

	c.SetCookie(&http.Cookie{
		Name:    "uin",
		Value:   strconv.FormatUint(uint64(uin), 10),
		Expires: time.Now().Add(ttl),
	})

	s.logger.Info("created new session", "uin", uin)

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (s *Server) handleRegister(c *echo.Context) error {
	if GetHasAuth(c) {
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}

	data := map[string]any{}
	data["missingValues"] = QueryParamBool(c, "missingValues")
	data["usernameTooLong"] = QueryParamBool(c, "usernameTooLong")
	data["badEmail"] = QueryParamBool(c, "badEmail")
	data["badPasswordLen"] = QueryParamBool(c, "badPasswordLen")
	data["passwordMismatch"] = QueryParamBool(c, "passwordMismatch")
	data["serverError"] = QueryParamBool(c, "serverError")
	data["usernameNotUnique"] = QueryParamBool(c, "usernameNotUnique")
	data["emailNotUnique"] = QueryParamBool(c, "emailNotUnique")

	return c.Render(http.StatusOK, "register.jet", data)
}

func (s *Server) handleRegisterAction(c *echo.Context) error {
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm-password")

	// Verify if all values are present
	if username == "" || email == "" || password == "" || confirmPassword == "" {
		return c.Redirect(http.StatusSeeOther, "/register?missingValues=1")
	}

	if len(username) > 24 {
		return c.Redirect(http.StatusSeeOther, "/register?usernameTooLong=1")
	}

	if !VerifyEmail(email) {
		return c.Redirect(http.StatusSeeOther, "/register?badEmail=1")
	}

	if err := PasswordMeetsRequirements(password); err != nil {
		s.logger.Info("user's password doesn't meet requirements", "err", err)
		return c.Redirect(http.StatusSeeOther, "/register?badPasswordLen=1")
	}

	if password != confirmPassword {
		return c.Redirect(http.StatusSeeOther, "/register?passwordMismatch=1")
	}

	// Create all of the required password hashes
	pwdHash, err := argon2.HashPassword(password)
	if err != nil {
		s.logger.Error("failed to argon2 hash password", "err", err)
		return c.Redirect(http.StatusSeeOther, "/register?serverError=1")
	}

	ggAncientHash := gg.GGAncientLoginHash(password, utils.GetSeed())
	gg32Hash := gg.GG32LoginHash(password, utils.GetSeed())
	ggSha1Hash := gg.GGSHA1LoginHash(password, utils.GetSeed())

	newUin, err := s.db.Users.CreateUserNew(username, email, pwdHash, ggAncientHash, gg32Hash, ggSha1Hash)
	if err != nil {
		// Check if it's an unique value constraint violation
		// TODO: Move this check to the function call
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "gguser_name_unique":
				s.logger.Warn("username was not unique", "username", username)
				return c.Redirect(http.StatusSeeOther, "/register?usernameNotUnique=1")
			case "gguser_email_unique":
				s.logger.Warn("email was not unique", "email", email)
				return c.Redirect(http.StatusSeeOther, "/register?emailNotUnique=1")
			}
		}

		s.logger.Error("failed to create new user", "err", err)
		return c.Redirect(http.StatusSeeOther, "/register?serverError=1")
	}

	s.logger.Info("new user registered!", "username", username, "uin", newUin)

	return c.Redirect(http.StatusSeeOther, "/login?registerSuccess=1")
}
