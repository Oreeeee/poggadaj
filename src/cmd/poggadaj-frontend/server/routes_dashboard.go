// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package server

import (
	"errors"
	"net/http"

	"codeberg.org/or3e/poggadaj/internal/security/argon2"
	"codeberg.org/or3e/poggadaj/internal/security/gg"
	"codeberg.org/or3e/poggadaj/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v5"
)

func (s *Server) handleDashboard(c *echo.Context) error {
	data := map[string]any{}

	// Get data for the current user
	uin := GetUin(c)
	if uin == 0 {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userData, err := s.db.GetUserDataByUin(uin)
	if err != nil {
		s.logger.Error("failed to get data for user", "uin", uin, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data["user"] = userData
	data["emailChanged"] = QueryParamBool(c, "emailChanged")

	return c.Render(http.StatusOK, "dashboard.jet", data)
}

func (s *Server) handleChangePassword(c *echo.Context) error {
	data := map[string]any{}
	data["missingValues"] = QueryParamBool(c, "missingValues")
	data["badOriginalPassword"] = QueryParamBool(c, "badOriginalPassword")
	data["badPasswordLen"] = QueryParamBool(c, "badPasswordLen")
	data["passwordMismatch"] = QueryParamBool(c, "passwordMismatch")
	data["serverError"] = QueryParamBool(c, "serverError")

	return c.Render(http.StatusOK, "changepass.jet", data)
}

func (s *Server) handleChangePasswordAction(c *echo.Context) error {
	currentPassword := c.FormValueOr("currentPassword", "")
	newPassword := c.FormValueOr("newPassword", "")
	confirmPassword := c.FormValueOr("confirmPassword", "")

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		return c.Redirect(http.StatusSeeOther, "/dashboard/changePassword?missingValues=1")
	}

	uin := GetUin(c)
	if uin == 0 {
		return c.NoContent(http.StatusUnauthorized)
	}

	// Check if the current password is correct
	originalPassword, err := s.db.GetUserPasswordByUin(uin)
	if err != nil {
		s.logger.Error("failed to get user password from the database", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	match, err := argon2.ComparePasswords(currentPassword, originalPassword)
	if err != nil {
		s.logger.Error("failed to compare passwords", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if !match {
		return c.Redirect(http.StatusSeeOther, "/dashboard/changePassword?badOriginalPassword=1")
	}

	// Check if the passwords match
	if newPassword != confirmPassword {
		return c.Redirect(http.StatusSeeOther, "/dashboard/changePassword?passwordMismatch=1")
	}

	// Check if the requirements are met
	if err := PasswordMeetsRequirements(newPassword); err != nil {
		s.logger.Info("password doesn't meet requirements", "err", err)
		return c.Redirect(http.StatusSeeOther, "/dashboard/changePassword?badPasswordLen=1")
	}

	// Generate the new passwords
	pwdHash, err := argon2.HashPassword(newPassword)
	if err != nil {
		s.logger.Error("failed to argon2 hash password", "err", err)
		return c.Redirect(http.StatusSeeOther, "/dashboard/changePassword?serverError=1")
	}

	ggAncientHash := gg.GGAncientLoginHash(newPassword, utils.GetSeed())
	gg32Hash := gg.GG32LoginHash(newPassword, utils.GetSeed())
	ggSha1Hash := gg.GGSHA1LoginHash(newPassword, utils.GetSeed())

	err = s.db.UpdateUserPassword(uin, pwdHash, ggAncientHash, gg32Hash, ggSha1Hash)
	if err != nil {
		s.logger.Error("failed to change user's password", "uin", uin, "err", err)
		return c.Redirect(http.StatusSeeOther, "/dashboard/changePassword?serverError=1")
	}

	// Log out the user and prompt to log back in
	// TODO: This should invalidate all active sessions
	err = LogoutUser(c, s)
	if err != nil {
		if errors.Is(err, LogoutWithoutToken) {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.Redirect(http.StatusSeeOther, "/login?passwordChanged=1")
}

func (s *Server) handleChangeEmail(c *echo.Context) error {
	data := map[string]any{}

	uin := GetUin(c)
	if uin == 0 {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userData, err := s.db.GetUserDataByUin(uin)
	if err != nil {
		s.logger.Error("failed to get data for user", "uin", uin, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data["currentEmail"] = userData.Email

	data["badPassword"] = QueryParamBool(c, "badPassword")
	data["badEmail"] = QueryParamBool(c, "badEmail")
	data["missingValues"] = QueryParamBool(c, "missingValues")
	data["emailNotUnique"] = QueryParamBool(c, "emailNotUnique")

	return c.Render(http.StatusOK, "changeemail.jet", data)
}

func (s *Server) handleChangeEmailAction(c *echo.Context) error {
	newEmail := c.FormValueOr("newEmail", "")
	password := c.FormValueOr("password", "")

	if newEmail == "" || password == "" {
		return c.Redirect(http.StatusSeeOther, "/dashboard/changeEmail?missingValues=1")
	}

	if !VerifyEmail(newEmail) {
		return c.Redirect(http.StatusSeeOther, "/dashboard/changeEmail?badEmail=1")
	}

	uin := GetUin(c)
	if uin == 0 {
		return c.NoContent(http.StatusUnauthorized)
	}

	originalPassword, err := s.db.GetUserPasswordByUin(uin)
	if err != nil {
		s.logger.Error("failed to get user password from the database", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	match, err := argon2.ComparePasswords(password, originalPassword)
	if err != nil {
		s.logger.Error("failed to compare passwords", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if !match {
		return c.Redirect(http.StatusSeeOther, "/dashboard/changeEmail?badPassword=1")
	}

	err = s.db.UpdateUserEmail(uin, newEmail)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			s.logger.Warn("email was not unique", "newEmail", newEmail)
			return c.Redirect(http.StatusSeeOther, "/dashboard/changeEmail?emailNotUnique=1")
		}

		s.logger.Error("failed to change user email", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard?emailChanged=1")
}

func (s *Server) handleLogout(c *echo.Context) error {
	err := LogoutUser(c, s)
	if err != nil {
		if errors.Is(err, LogoutWithoutToken) {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	// Now, redirect the user to the login page with the correct message
	return c.Redirect(http.StatusSeeOther, "/login?loggedOut=1")
}
