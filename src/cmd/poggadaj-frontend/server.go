// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"context"
	"log/slog"
	"net/http"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/utils/utilshttp"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Server struct {
	e      *echo.Echo
	ip     string
	mc     *MiddlewareController
	logger *log.Logger
}

func (s *Server) handleHome(c *echo.Context) error {
	return c.Render(http.StatusOK, "home.jet", nil)
}

func (s *Server) handleLogin(c *echo.Context) error {
	return c.Render(http.StatusOK, "login.jet", nil)
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

func NewServer(ip string, logger *log.Logger, renderer *TemplateRenderer) (*Server, error) {
	server := &Server{}

	server.e = echo.New()
	utilshttp.SetUpLogger(server.e, logger)

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
	server.e.GET("/", server.handleHome, server.mc.LanguageMiddleware)
	server.e.GET("/login", server.handleLogin, server.mc.LanguageMiddleware)
	server.e.GET("/download", server.handleDownloads, server.mc.LanguageMiddleware)

	return server, nil
}
