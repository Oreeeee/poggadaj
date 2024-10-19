package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

var DatabaseConn *pgxpool.Pool

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

func main() {
	dbconn, _ := GetDBConn()
	DatabaseConn = dbconn

	r := echo.New()
	r.HideBanner = true
	r.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	r.POST("/api/v1/register", registerUser)
	r.Logger.Fatal(
		r.Start(
			fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
		),
	)
}
