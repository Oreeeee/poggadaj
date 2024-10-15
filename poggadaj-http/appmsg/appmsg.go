package appmsg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func AppMSG_Handler(c *gin.Context) {
	ip := os.Getenv("GG_SERVICE_IP")
	port := os.Getenv("GG_SERVICE_PORT")
	c.String(http.StatusOK, fmt.Sprintf("0 0 %s:%s %s", ip, port, ip))
}
