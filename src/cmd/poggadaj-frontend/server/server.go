// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package server

import (
	"context"
	"log/slog"
	"os"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/cache"
	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/utils/utilshttp"
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
	server.e.GET("/dashboard", server.handleDashboard, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware, server.mc.RequireAuthMiddleware)
	server.e.GET("/dashboard/changePassword", server.handleChangePassword, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware, server.mc.RequireAuthMiddleware)
	server.e.POST("/dashboard/changePassword", server.handleChangePasswordAction, server.mc.HasAuthMiddleware, server.mc.RequireAuthMiddleware)
	server.e.GET("/dashboard/changeEmail", server.handleChangeEmail, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware, server.mc.RequireAuthMiddleware)
	server.e.POST("/dashboard/changeEmail", server.handleChangeEmailAction, server.mc.HasAuthMiddleware, server.mc.RequireAuthMiddleware)
	server.e.POST("/logout", server.handleLogout, server.mc.HasAuthMiddleware, server.mc.RequireAuthMiddleware)
	server.e.GET("/download", server.handleDownloads, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)
	server.e.GET("/connection-guide", server.handleConnectionGuide, server.mc.LanguageMiddleware, server.mc.HasAuthMiddleware)

	return server, nil
}
