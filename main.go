package main

import (
	"net/http"

	"github.com/TimotejKovacka/tiny-url/urlservice"
	"github.com/gin-gonic/gin"
)

type CreateRequestBody struct {
	LongURL string `json:"long_url"`
}

func main() {
	r := gin.Default()
	const BASE_URL = "https://tiny-url.com/"
	urlService := urlservice.NewURLService()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("create", func(c *gin.Context) {
		var reqBody CreateRequestBody

		if err := c.BindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		shortURL := BASE_URL + urlService.LongToShort(reqBody.LongURL)

		c.JSON(http.StatusCreated, gin.H{
			"short_url": shortURL,
		})
	})
	r.Run()
}
