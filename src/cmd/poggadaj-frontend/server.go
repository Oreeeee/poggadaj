// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/cache"
	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/security/argon2"
	"codeberg.org/or3e/poggadaj/internal/security/gg"
	"codeberg.org/or3e/poggadaj/internal/utils"
	"codeberg.org/or3e/poggadaj/internal/utils/utilshttp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Server struct {
	e      *echo.Echo
	ip     string
	mc     *MiddlewareController
	db     *database.Database
	cache  *cache.Cache
	logger *log.Logger
}

func (s *Server) handleHome(c *echo.Context) error {
	return c.Render(http.StatusOK, "home.jet", nil)
}

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
	uin, passwordHash, err := s.db.GetUserPasswordHashWithUin(username)

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

func (s *Server) handleResetPass(c *echo.Context) error {
	return c.Render(http.StatusOK, "resetpass.jet", nil)
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

	newUin, err := s.db.CreateUserNew(username, email, pwdHash, ggAncientHash, gg32Hash, ggSha1Hash)
	if err != nil {
		// Check if it's an unique value constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "gguser_name_key":
				s.logger.Warn("username was not unique", "username", username)
				return c.Redirect(http.StatusSeeOther, "/register?usernameNotUnique=1")
			case "gguser_email_key":
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

func (s *Server) handleDashboard(c *echo.Context) error {
	// TODO: once authentication middleware exists, insert the data as a param.
	// Or actually add some thing to TemplateArgs for this

	// Get data for the current user
	uin := GetUin(c)
	if uin == 0 {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	data, err := s.db.GetUserDataByUin(uin)
	if err != nil {
		s.logger.Error("failed to get data for user", "uin", uin, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

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

func (s *Server) handleDownloads(c *echo.Context) error {
	lang := GetLang(c)
	if lang == "" {
		s.logger.Warn("didn't get a value for lang, using en")
		lang = "en"
	}

	clients, err := s.db.GetClients(lang)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.Render(http.StatusOK, "downloads.jet", map[string]any{"Clients": clients})
}

func (s *Server) handleConnectionGuide(c *echo.Context) error {
	return c.Render(http.StatusOK, "connectionguide.jet", nil)
}

func (s *Server) Run() error {
	return s.e.Start(s.ip)
}

func NewServer(ip string, logger *log.Logger, renderer *TemplateRenderer, db *database.Database, cache *cache.Cache) (*Server, error) {
	server := &Server{}
	server.ip = ip
	server.db = db
	server.logger = logger
	server.cache = cache

	server.e = echo.New()
	utilshttp.SetUpLogger(server.e, server.logger)

	var err error
	server.mc, err = NewMiddlewareController(server)
	if err != nil {
		return nil, err
	}

	server.e.Renderer = renderer

	server.e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				server.e.Logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				server.e.Logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	// Ensure that the files directory exists
	err = os.Mkdir("files", 0777)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	server.e.Static("/static", "static")
	server.e.Static("/files", "files")
	server.e.GET("/", server.handleHome, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.GET("/login", server.handleLogin, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.POST("/login", server.handleLoginAction)
	server.e.GET("/resetpass", server.handleResetPass, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.GET("/register", server.handleRegister, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.POST("/register", server.handleRegisterAction)
	server.e.GET("/dashboard", server.handleDashboard, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)                     // TODO: Add authentication middleware
	server.e.GET("/dashboard/changePassword", server.handleChangePassword, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware) // TODO: Add authentication middleware
	server.e.POST("/dashboard/changePassword", server.handleChangePasswordAction, server.mc.HasAuthMiddleware)                        // TODO: Add authentication middleware
	server.e.POST("/logout", server.handleLogout, server.mc.HasAuthMiddleware)                                                        // TODO: Add authentication middleware
	server.e.GET("/download", server.handleDownloads, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.GET("/connection-guide", server.handleConnectionGuide, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)

	return server, nil
}
