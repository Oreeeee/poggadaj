package main

import (
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

	ip := "0.0.0.0:8080" // When running in Docker
	if len(os.Args) > 1 && os.Args[1] == "dockerless" {
		ip = "127.0.0.1:80"
	}

	log.Fatal(r.Run(ip))
}
