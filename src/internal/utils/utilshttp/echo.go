// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package utilshttp

import (
	"log/slog"

	"charm.land/log/v2"
	"github.com/labstack/echo/v5"
)

func SetUpLogger(e *echo.Echo, logger *log.Logger) {
	slogLogger := slog.New(logger)
	e.Logger = slogLogger
}
