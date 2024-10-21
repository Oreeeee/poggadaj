package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
)

var DatabaseConn *pgxpool.Pool
var Sessions []AuthorizedSession

func registerUser(c echo.Context) error {
	name := c.FormValue("name")
	password := c.FormValue("password")

	if !PasswordFitsRestrictions(password) {
		return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "Password doesn't fit constraints"})
	}
	uin, err := CreateUser(name, password)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "Unknown error when creating user"})
	}

	return c.JSON(http.StatusOK, &RegisterResponse{UIN: uin})
}

func loginUser(c echo.Context) error {
	name := c.FormValue("name")
	password := c.FormValue("password")
	passwordHash, _ := GetUserPasswordHash(name)
	passwordMatch, _ := ComparePasswords(password, passwordHash)
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

func gg32ChangePassword(c echo.Context) error {
	sessionValid, username := ValidateSession(c)
	if !sessionValid {
		return c.String(http.StatusUnauthorized, "")
	}

	password := c.FormValue("password")
	if !PasswordFitsRestrictions(password) {
		return c.String(http.StatusBadRequest, "Password doesn't fit constraints")
	}
	if UpdateUserGG32(username, password, 0xFFFFFFFF) != nil {
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
	r.GET("/api/v1/register", registerUser)
	r.GET("/api/v1/login", loginUser)
	r.GET("/api/v1/gg32-changepwd", gg32ChangePassword)
	r.GET("/api/v1/is-authenticated", isAuthenticated)
	r.GET("/api/v1/user-data", userData)
	r.Logger.Fatal(
		r.Start(
			fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
		),
	)
}
