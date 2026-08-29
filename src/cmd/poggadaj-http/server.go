// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"charm.land/log/v2"
	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/logging"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"codeberg.org/or3e/poggadaj/internal/utils/utilshttp"
	"github.com/labstack/echo/v5"
)

type Server struct {
	e      *echo.Echo
	db     *database.Database
	logger *log.Logger
}

func (server *Server) AppMSG_Handler(c *echo.Context) error {
	ip := os.Getenv("GG_SERVICE_IP")
	port := os.Getenv("GG_SERVICE_PORT")
	return c.String(http.StatusOK, fmt.Sprintf("0 0 %s:%s %s", ip, port, ip))
}

func (server *Server) buildResponse(bannerType int) string {
	ads := server.db.GetAds(bannerType)
	// TODO: Give the correct dimensions depending on the banner type once Wayback Machine goes back online
	imageFmt := "<img src=\"%s\" />"
	response := ""

	if len(ads) == 0 {
		// Nothing really to choose from
		return response
	}

	// Select a random ad
	ad := ads[rand.Intn(len(ads))]

	// Build an image response if we got an image ad
	if ad.AdType == structs.ADTYPE_IMAGE {
		// TODO: Add image support
		return response
		imageUrl := ad.Image.String
		response = fmt.Sprintf(imageFmt, imageUrl)
	} else {
		response = ad.Html.String
	}

	return response
}

func (server *Server) GetMainBanner(c *echo.Context) error {
	return c.String(http.StatusOK, server.buildResponse(BANNERTYPE_MAIN))
}

func (server *Server) GetSmallBanner(c *echo.Context) error {
	return c.String(http.StatusOK, server.buildResponse(BANNERTYPE_SMALL))
}

func (server *Server) GetBanner(c *echo.Context) error {
	return c.String(http.StatusOK, server.buildResponse(BANNERTYPE_BANNER))
}

func (server *Server) Run() {
	server.e.Start(
		fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
	)
}

func NewServer(dbCfg *database.DatabaseConfig) (*Server, error) {
	server := &Server{}
	server.logger = logging.NewLogger()
	server.e = echo.New()
	utilshttp.SetUpLogger(server.e, server.logger)

	db, err := database.NewDatabase(dbCfg, logging.NewLoggerWithPrefix("database"))
	if err != nil {
		return nil, err
	}
	server.db = db

	// appmsg.gadu-gadu.pl
	server.e.GET("/appsvc/appmsg4.asp",
		server.AppMSG_Handler,
	)
	server.e.GET("/appsvc/appmsg2.asp",
		func(c *echo.Context) error {
			ip := os.Getenv("GG_SERVICE_IP")
			port := os.Getenv("GG_SERVICE_PORT")
			return c.String(http.StatusOK, fmt.Sprintf("0 %s:%s %s", ip, port, ip))
		},
	)
	server.e.GET("/appsvc/appmsg.asp",
		func(c *echo.Context) error {
			ip := os.Getenv("GG_SERVICE_IP")
			port := os.Getenv("GG_SERVICE_PORT")
			return c.String(http.StatusOK, fmt.Sprintf("0 1 0 %s:%s %s %s", ip, port, ip, ip))
		},
	)

	// adserver.gadu-gadu.pl
	// TODO: Implement different responses depending on the endpoint
	// TODO: Make the responses configurable
	server.e.GET("/getmainbanner.asp",
		server.GetMainBanner,
	)
	server.e.GET("/smallbanner.asp",
		server.GetSmallBanner,
	)
	server.e.GET("/getbanner.asp",
		server.GetBanner,
	)

	return server, nil
}
