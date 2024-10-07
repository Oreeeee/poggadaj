package appmsg

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AppMSG_Handler(c *gin.Context) {
	// TODO: Dynamically return the IP addresses
	c.String(http.StatusOK, "26679 0 127.0.0.1:8074 127.0.0.1")
}
