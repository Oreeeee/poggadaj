// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package appmsg

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
)

func AppMSG_Handler(c *echo.Context) error {
	ip := os.Getenv("GG_SERVICE_IP")
	port := os.Getenv("GG_SERVICE_PORT")
	return c.String(http.StatusOK, fmt.Sprintf("0 0 %s:%s %s", ip, port, ip))
}
