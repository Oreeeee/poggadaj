package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetMainBanner_Handler(c *gin.Context) {
	c.String(http.StatusOK, "Hello from poggadaj-HTTP")
}
