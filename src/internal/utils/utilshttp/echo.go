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
