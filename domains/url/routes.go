package url

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateRequestBody struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

type RoutesConfig struct {
	BASE_URL string
	DB       *gorm.DB
}

func RegisterRoutes(r *gin.Engine, config *RoutesConfig) {
	urlStorage := NewURLStorage(config.DB)
	urlService := NewURLService(urlStorage)

	r.POST("/create", func(c *gin.Context) {
		var reqBody CreateRequestBody

		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		shortURL, err := urlService.LongToShort(reqBody.LongURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"short_url": config.BASE_URL + shortURL,
		})
	})

	r.GET("/:urlHash", func(c *gin.Context) {
		urlHash := c.Param("urlHash")

		longURL, err := urlService.ShortToLong(urlHash)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
			return
		}

		c.Redirect(http.StatusMovedPermanently, longURL)
	})
}
