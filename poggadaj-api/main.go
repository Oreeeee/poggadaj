package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

var DatabaseConn *pgxpool.Pool
var Sessions []AuthorizedSession

func registerUser(c echo.Context) error {
	name := c.FormValue("name")
	password := c.FormValue("password")

	if len(password) > 64 || len(password) < 8 {
		return c.JSON(http.StatusBadRequest, &RegisterResponse{Error: "Password doesn't fit these constraints: >7<64"})
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

func main() {
	dbconn, _ := GetDBConn()
	DatabaseConn = dbconn

	r := echo.New()
	r.HideBanner = true
	r.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	r.POST("/api/v1/register", registerUser)
	r.POST("/api/v1/login", loginUser)
	r.Logger.Fatal(
		r.Start(
			fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
		),
	)
}
