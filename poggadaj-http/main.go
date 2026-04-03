// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"poggadaj_http/appmsg"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

var DatabaseConn *pgxpool.Pool

func main() {
	dbconn, _ := GetDBConn()
	DatabaseConn = dbconn

	r := echo.New()

	// appmsg.gadu-gadu.pl
	r.GET("/appsvc/appmsg4.asp",
		appmsg.AppMSG_Handler,
	)
	r.GET("/appsvc/appmsg2.asp",
		appmsg.AppMSG_Handler,
	)
	r.GET("/appsvc/appmsg.asp",
		func(c *echo.Context) error {
			ip := os.Getenv("GG_SERVICE_IP")
			port := os.Getenv("GG_SERVICE_PORT")
			return c.String(http.StatusOK, fmt.Sprintf("0 1 0 %s:%s %s %s", ip, port, ip, ip))
		},
	)

	// adserver.gadu-gadu.pl
	// TODO: Implement different responses depending on the endpoint
	// TODO: Make the responses configurable
	r.GET("/getmainbanner.asp",
		GetMainBanner,
	)
	r.GET("/smallbanner.asp",
		GetSmallBanner,
	)
	r.GET("/getbanner.asp",
		GetBanner,
	)

	log.Fatal(
		r.Start(
			fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
		),
	)
}
