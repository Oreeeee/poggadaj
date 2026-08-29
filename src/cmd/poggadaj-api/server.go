// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/logging"
	"codeberg.org/or3e/poggadaj/internal/security/argon2"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"codeberg.org/or3e/poggadaj/internal/utils/utilshttp"
	"github.com/labstack/echo/v5"
)

type Server struct {
	e      *echo.Echo
	db     *database.Database
	logger *log.Logger
}

func (server *Server) registerUser(c *echo.Context) error {
	regBody := structs.RegisterRequest{}
	bodyErr := json.NewDecoder(c.Request().Body).Decode(&regBody)
	if bodyErr != nil {
		return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "Failed to unmarshal register request"})
	}

	uin, err := server.db.CreateUser(regBody)
	if err != nil {
		fmt.Println(err)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"gguser_name_key\"") {
			return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "User with this name already exists"})
		}
		return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "Unknown error when creating user"})
	}

	return c.JSON(http.StatusOK, &RegisterResponse{UIN: uin})
}

func (server *Server) loginUser(c *echo.Context) error {
	name := c.FormValue("name")
	password := c.FormValue("password")
	passwordHash, _ := server.db.GetUserPasswordHash(name)
	passwordMatch, _ := argon2.ComparePasswords(password, passwordHash)
	if passwordMatch {
		// Add the session to the authorized session list
		authSession := GenerateAuthorizedSession(name)
		Sessions = append(Sessions, authSession)

		// Create an auth cookie for the client
		authCookie := http.Cookie{
			Name:    "Auth",
			Value:   authSession.AuthCookie,
			Expires: authSession.Expires,
		}
		c.SetCookie(&authCookie)

		// Create username cookie
		usernameCookie := http.Cookie{
			Name:    "Username",
			Value:   name,
			Expires: authSession.Expires,
		}
		c.SetCookie(&usernameCookie)

		return c.String(http.StatusOK, "")
	}
	return c.String(http.StatusUnauthorized, "")
}

func (server *Server) changePassword(c *echo.Context) error {
	sessionValid, username := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}

	body := structs.ChangePasswordRequest{}
	bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
	if bodyErr != nil {
		return c.String(http.StatusBadRequest, "Failed to unmarshal ChangePasswordRequest")
	}

	err := server.db.UpdateUserPassword(username, body)
	if err != nil {
		if strings.Contains(err.Error(), "Wrong password type") {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusInternalServerError, "")
	}
	return c.String(http.StatusOK, "")
}

func (server *Server) changeClientsPassword(c *echo.Context) error {
	sessionValid, username := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}

	body := structs.ChangePasswordRequest{}
	bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
	if bodyErr != nil {
		return c.String(http.StatusBadRequest, "Failed to unmarshal ChangePasswordRequest")
	}

	err1 := server.db.UpdateAncientPassword(username, body.Password)
	if err1 != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	err2 := server.db.UpdateGG32Password(username, body.Password)
	if err2 != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	err3 := server.db.UpdateSHA1Password(username, body.Password)
	if err3 != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	return c.String(http.StatusOK, "")
}

func (server *Server) isAuthenticated(c *echo.Context) error {
	sessionValid, _ := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}
	return c.String(http.StatusOK, "")
}

func (server *Server) userData(c *echo.Context) error {
	sessionValid, username := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}
	uin, joined, err := server.db.GetUserData(username)
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusBadRequest, "{}")
	}
	return c.JSON(http.StatusOK, UserDataResponse{
		UIN:    uin,
		Joined: joined,
	})
}

func (server *Server) Run() {
	server.e.Start(fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")))
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

	//server.e.Use(middleware.CORS()) // TODO: Configure
	server.e.POST("/api/v1/register", server.registerUser)
	server.e.GET("/api/v1/login", server.loginUser)
	server.e.POST("/api/v1/changepassword", server.changePassword)
	server.e.POST("/api/v1/chgclpwd", server.changeClientsPassword)
	server.e.GET("/api/v1/is-authenticated", server.isAuthenticated)
	server.e.GET("/api/v1/user-data", server.userData)

	return server, nil
}
