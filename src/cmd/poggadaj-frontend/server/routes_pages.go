// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (s *Server) handleHome(c *echo.Context) error {
	return c.Render(http.StatusOK, "home.jet", nil)
}

func (s *Server) handleDownloads(c *echo.Context) error {
	lang := GetLang(c)
	if lang == "" {
		s.logger.Warn("didn't get a value for lang, using en")
		lang = "en"
	}

	clients, err := s.db.Clients.GetClients(lang)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.Render(http.StatusOK, "downloads.jet", map[string]any{"Clients": clients})
}

func (s *Server) handleConnectionGuide(c *echo.Context) error {
	return c.Render(http.StatusOK, "connectionguide.jet", nil)
}

func (s *Server) handleResetPass(c *echo.Context) error {
	return c.Render(http.StatusOK, "resetpass.jet", nil)
}
