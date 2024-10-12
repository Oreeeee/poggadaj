package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"poggadaj_http/appmsg"
)

var DatabaseConn *pgxpool.Pool

func main() {
	dbconn, _ := GetDBConn()
	DatabaseConn = dbconn

	r := gin.Default()

	// appmsg.gadu-gadu.pl
	r.GET("/appsvc/appmsg4.asp",
		appmsg.AppMSG_Handler,
	)

	// adserver.gadu-gadu.pl
	// TODO: Implement different responses depending on the endpoint
	// TODO: Make the responses configurable
	r.GET("/getmainbanner.asp",
		GetMainBanner_Handler,
	)
	r.GET("/smallbanner.asp",
		GetMainBanner_Handler,
	)
	r.GET("/getbanner.asp",
		GetMainBanner_Handler,
	)

	log.Fatal(
		r.Run(
			fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
		),
	)
}
