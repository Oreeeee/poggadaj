package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetMainBanner(c *gin.Context) {
	ads := GetAds(BANNERTYPE_MAIN)
	c.String(http.StatusOK, ads[0].Html.String)
}

func GetSmallBanner(c *gin.Context) {
	ads := GetAds(BANNERTYPE_SMALL)
	c.String(http.StatusOK, ads[0].Html.String)
}

func GetBanner(c *gin.Context) {
	ads := GetAds(BANNERTYPE_BANNER)
	c.String(http.StatusOK, ads[0].Html.String)
}
