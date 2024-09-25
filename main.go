package main

import (
	"context"
	"net/http"

	"github.com/TimotejKovacka/tiny-url/domains/url"
	"github.com/TimotejKovacka/tiny-url/internal/database"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

const (
	BASE_URL        = "https://tiny-url.com/"
	SERVICE_NAME    = "tiny-url-service"
	SERVICE_VERSION = "1.0.0"
	OTEL_COLLECTOR  = "localhost:4317"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(OTEL_COLLECTOR),
	)
	if err != nil {
		return nil, err
	}

	resource, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(SERVICE_NAME),
			semconv.ServiceVersionKey.String(SERVICE_VERSION),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	// Initialize Zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Initialize tracer
	tp, err := initTracer()
	if err != nil {
		sugar.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			sugar.Errorf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize database
	dbConfig := database.Config{
		Host:     "localhost",
		User:     "admin",
		Password: "admin",
		DBName:   "test_db",
		Port:     "5432",
	}
	db, err := database.ConnectDB(dbConfig, sugar)
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Gin
	r := gin.New()
	r.Use(otelgin.Middleware(SERVICE_NAME))
	r.Use(gin.Recovery())

	// Add a middleware to inject the logger into the context
	r.Use(func(c *gin.Context) {
		c.Set("logger", sugar)
		c.Next()
	})

	r.GET("/ping", func(c *gin.Context) {
		logger := c.MustGet("logger").(*zap.SugaredLogger)
		logger.Info("Received ping request")
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	url.RegisterRoutes(r, &url.RoutesConfig{
		DB:       db,
		BASE_URL: BASE_URL,
		Logger:   sugar,
		Tracer:   otel.Tracer("url-service"),
	})

	// Start the server
	sugar.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		sugar.Fatalf("Failed to start server: %v", err)
	}
}
