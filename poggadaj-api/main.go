// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package main

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"poggadaj-shared/security/argon2"
	"strings"
)

var DatabaseConn *pgxpool.Pool
var Sessions []AuthorizedSession

func registerUser(c echo.Context) error {
	regBody := RegisterRequest{}
	bodyErr := json.NewDecoder(c.Request().Body).Decode(&regBody)
	if bodyErr != nil {
		return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "Failed to unmarshal register request"})
	}

	uin, err := CreateUser(regBody)
	if err != nil {
		fmt.Println(err)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"gguser_name_key\"") {
			return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "User with this name already exists"})
		}
		return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "Unknown error when creating user"})
	}

	return c.JSON(http.StatusOK, &RegisterResponse{UIN: uin})
}

func loginUser(c echo.Context) error {
	name := c.FormValue("name")
	password := c.FormValue("password")
	passwordHash, _ := GetUserPasswordHash(name)
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

func changePassword(c echo.Context) error {
	sessionValid, username := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}

	body := ChangePasswordRequest{}
	bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
	if bodyErr != nil {
		return c.String(http.StatusBadRequest, "Failed to unmarshal ChangePasswordRequest")
	}

	err := UpdateUserPassword(username, body)
	if err != nil {
		if strings.Contains(err.Error(), "Wrong password type") {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusInternalServerError, "")
	}
	return c.String(http.StatusOK, "")
}

func changeClientsPassword(c echo.Context) error {
	sessionValid, username := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}

	body := ChangePasswordRequest{}
	bodyErr := json.NewDecoder(c.Request().Body).Decode(&body)
	if bodyErr != nil {
		return c.String(http.StatusBadRequest, "Failed to unmarshal ChangePasswordRequest")
	}

	err1 := UpdateAncientPassword(username, body.Password)
	if err1 != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	err2 := UpdateGG32Password(username, body.Password)
	if err2 != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	err3 := UpdateSHA1Password(username, body.Password)
	if err3 != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	return c.String(http.StatusOK, "")
}

func isAuthenticated(c echo.Context) error {
	sessionValid, _ := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}
	return c.String(http.StatusOK, "")
}

func userData(c echo.Context) error {
	sessionValid, username := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}
	uin, joined, err := GetUserData(username)
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusBadRequest, "{}")
	}
	return c.JSON(http.StatusOK, UserDataResponse{
		UIN:    uin,
		Joined: joined,
	})
}

func main() {
	dbconn, _ := GetDBConn()
	DatabaseConn = dbconn

	r := echo.New()
	r.HideBanner = true
	r.Use(middleware.CORS()) // TODO: Configure
	r.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	r.POST("/api/v1/register", registerUser)
	r.GET("/api/v1/login", loginUser)
	r.POST("/api/v1/changepassword", changePassword)
	r.POST("/api/v1/chgclpwd", changeClientsPassword)
	r.GET("/api/v1/is-authenticated", isAuthenticated)
	r.GET("/api/v1/user-data", userData)
	r.Logger.Fatal(
		r.Start(
			fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
		),
	)
}
