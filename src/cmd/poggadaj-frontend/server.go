// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/cache"
	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/security/argon2"
	"codeberg.org/or3e/poggadaj/internal/utils/utilshttp"
	"github.com/jackc/pgx/v5"
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

	failedLogin := c.QueryParamOr("fail", "0")
	if failedLogin == "1" {
		data["showLoginFail"] = true
	}

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

func (s *Server) handleDashboard(c *echo.Context) error {
	// TODO: once authentication middleware exists, insert the data as a param.
	// Or actually add some thing to TemplateArgs for this

	// Get data for the current user
	var uin uint
	if uinRaw := c.Get("poggadaj-uin"); uinRaw != nil {
		uin = uinRaw.(uint)
	} else {
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
	return c.Render(http.StatusOK, "changepass.jet", nil)
}

func (s *Server) handleDownloads(c *echo.Context) error {
	clients := []HtmlClient{
		{
			Name:               "Gadu-Gadu 6.0",
			DescriptionI18nTag: "gg60-description",
			ImageUrl:           "../static/gg60.png",
			DownloadUrl:        "https://example.com",
		},
		{
			Name:               "Gadu-Gadu 7.7",
			DescriptionI18nTag: "gg77-description",
			ImageUrl:           "../static/gg77.png",
			DownloadUrl:        "https://example.com",
		},
	}
	return c.Render(http.StatusOK, "downloads.jet", map[string]any{"Clients": clients})
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

	server.e.Static("/static", "static")
	server.e.GET("/", server.handleHome, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.GET("/login", server.handleLogin, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.POST("/login", server.handleLoginAction)
	server.e.GET("/dashboard", server.handleDashboard, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware) // TODO: Add authentication middleware
	server.e.GET("/dashboard/changePassword", server.handleChangePassword, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.GET("/download", server.handleDownloads, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)

	return server, nil
}
