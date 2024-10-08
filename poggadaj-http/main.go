package main

import (
	"github.com/gin-gonic/gin"
	"log"
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
	log.Fatal(r.Run("127.0.0.1:80"))
}
