package url

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreateRequestBody struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

type RoutesConfig struct {
	BASE_URL string
	DB       *gorm.DB
	Tracer   trace.Tracer
	Logger   *zap.SugaredLogger
}

func RegisterRoutes(r *gin.Engine, config *RoutesConfig) {
	urlStorage := NewURLStorage(config.DB, config.Tracer, config.Logger)
	urlService := NewURLService(urlStorage, config.Tracer, config.Logger)

	r.POST("/create", func(c *gin.Context) {
		ctx, span := config.Tracer.Start(c.Request.Context(), "CreateShortURL")
		defer span.End()

		var reqBody CreateRequestBody

		if err := c.ShouldBindJSON(&reqBody); err != nil {
			config.Logger.Errorw("Invalid request body", "error", err)
			span.SetAttributes(attribute.String("error", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		shortURL, err := urlService.LongToShort(ctx, reqBody.LongURL)
		if err != nil {
			config.Logger.Errorw("Failed to create short URL", "error", err)
			span.SetAttributes(attribute.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
			return
		}

		config.Logger.Infow("Created new short URL", "short_url", shortURL, "long_url", reqBody.LongURL)
		span.SetAttributes(
			attribute.String("short_url", shortURL),
			attribute.String("long_url", reqBody.LongURL),
		)

		c.JSON(http.StatusCreated, gin.H{
			"short_url": config.BASE_URL + shortURL,
		})
	})

	r.GET("/:urlHash", func(c *gin.Context) {
		ctx, span := config.Tracer.Start(c.Request.Context(), "GetLongURL")
		defer span.End()

		urlHash := c.Param("urlHash")
		span.SetAttributes(attribute.String("url_hash", urlHash))

		longURL, err := urlService.ShortToLong(ctx, urlHash)
		if err != nil {
			config.Logger.Errorw("URL not found", "url_hash", urlHash)
			span.SetAttributes(attribute.String("error", err.Error()))
			c.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
			return
		}

		config.Logger.Infow("Redirecting", "url_hash", urlHash, "long_url", longURL)
		span.SetAttributes(attribute.String("long_url", longURL))

		c.Redirect(http.StatusMovedPermanently, longURL)
	})
}
