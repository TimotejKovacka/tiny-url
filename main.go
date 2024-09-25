package main

import (
	"net/http"

	"github.com/TimotejKovacka/tiny-url/domains/url"
	"github.com/TimotejKovacka/tiny-url/internal/database"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
)

type CreateRequestBody struct {
	LongURL string `json:"long_url"`
}

func startServer(r *gin.Engine) {
	if err := r.Run(":8080"); err != nil {
		panic("failed to start server")
	}
}

func main() {
	const BASE_URL = "https://tiny-url.com/"
	db := database.ConnectDB()
	r := gin.New()
	r.Use(logger.SetLogger())
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	url.RegisterRoutes(r, &url.RoutesConfig{
		DB:       db,
		BASE_URL: BASE_URL,
	})

	startServer(r)
}
