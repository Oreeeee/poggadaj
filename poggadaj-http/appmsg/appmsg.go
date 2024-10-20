package appmsg

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

func AppMSG_Handler(c echo.Context) error {
	ip := os.Getenv("GG_SERVICE_IP")
	port := os.Getenv("GG_SERVICE_PORT")
	return c.String(http.StatusOK, fmt.Sprintf("0 0 %s:%s %s", ip, port, ip))
}
