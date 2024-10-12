package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetMainBanner(c *gin.Context) {
	c.String(http.StatusOK, "Hello from poggadaj-HTTP")
}

func GetSmallBanner(c *gin.Context) {
	c.String(http.StatusOK, "Hello from poggadaj-HTTP")
}

func GetBanner(c *gin.Context) {
	c.String(http.StatusOK, "Hello from poggadaj-HTTP")
}
