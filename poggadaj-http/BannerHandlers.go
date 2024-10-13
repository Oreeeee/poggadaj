package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
)

func buildResponse(bannerType int) string {
	ads := GetAds(bannerType)
	// TODO: Give the correct dimensions depending on the banner type once Wayback Machine goes back online
	imageFmt := "<img src=\"%s\" />"
	response := ""

	if len(ads) == 0 {
		// Nothing really to choose from
		return response
	}

	// Select a random ad
	ad := ads[rand.Intn(len(ads))]

	// Build an image response if we got an image ad
	if ad.AdType == ADTYPE_IMAGE {
		// TODO: Add image support
		return response
		imageUrl := ad.Image.String
		response = fmt.Sprintf(imageFmt, imageUrl)
	} else {
		response = ad.Html.String
	}

	return response
}

func GetMainBanner(c *gin.Context) {
	c.String(http.StatusOK, buildResponse(BANNERTYPE_MAIN))
}

func GetSmallBanner(c *gin.Context) {
	c.String(http.StatusOK, buildResponse(BANNERTYPE_SMALL))
}

func GetBanner(c *gin.Context) {
	c.String(http.StatusOK, buildResponse(BANNERTYPE_BANNER))
}
