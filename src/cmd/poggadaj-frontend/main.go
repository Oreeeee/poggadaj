// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"context"
	"log/slog"
	"net/http"

	"codeberg.org/or3e/poggadaj/internal/logging"
	"codeberg.org/or3e/poggadaj/internal/utils/utilshttp"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	logger := logging.NewLogger()

	e := echo.New()
	utilshttp.SetUpLogger(e, logger)

	translator, err := NewTranslator("i18n")
	if err != nil {
		panic(err)
	}

	e.Renderer, err = NewTemplateRenderer("./views", true, translator)
	if err != nil {
		panic(err)
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				e.Logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				e.Logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	e.Static("/static", "static")

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "home.jet", nil)
	}, LanguageMiddleware)

	e.GET("/login", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "login.jet", nil)
	}, LanguageMiddleware)

	e.GET("/download", func(c *echo.Context) error {
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
	}, LanguageMiddleware)

	if err := e.Start(":3000"); err != nil {
		e.Logger.Error("shutting down the server", "error", err)
	}
}
