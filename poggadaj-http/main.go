package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"poggadaj_http/appmsg"
)

func main() {
	r := gin.Default()

	// appmsg.gadu-gadu.pl
	r.GET("/appsvc/appmsg4.asp",
		appmsg.AppMSG_Handler,
	)

	// adserver.gadu-gadu.pl
	r.GET("/getmainbanner.asp",
		GetMainBanner_Handler,
	)

	log.Fatal(
		r.Run(
			fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDRESS"), os.Getenv("LISTEN_PORT")),
		),
	)
}
