package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/TimotejKovacka/tiny-url/config"
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
	SERVICE_NAME    = "tiny-url-service"
	SERVICE_VERSION = "1.0.0"
)

func initTracer(config *config.ConfigYaml) (*sdktrace.TracerProvider, error) {
	var OTEL_COLLECTOR = config.Otel.Host + ":" + config.Otel.Port

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
	var (
		configFile string
	)
	flag.StringVar(&configFile, "c", "", "Configuration file path")
	flag.StringVar(&configFile, "config", "", "Configuration file path")

	flag.Usage = usage
	flag.Parse()

	// set default parameters.
	config, err := config.LoadConfig(configFile)
	if err != nil {
		log.Printf("Load yaml config file error: '%v'", err)

		return
	}

	// Initialize Zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Initialize tracer
	tp, err := initTracer(config)
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
		Host:     config.DB.Host,
		Port:     config.DB.Port,
		User:     config.DB.Username,
		Password: config.DB.Password,
		DBName:   config.DB.DBName,
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
		BASE_URL: config.Server.Host + ":" + config.Server.Port,
		Logger:   sugar,
		Tracer:   otel.Tracer("url-service"),
	})

	// Start the server
	sugar.Info("Starting server on :8080")
	if err := r.Run(":" + config.Server.Port); err != nil {
		sugar.Fatalf("Failed to start server: %v", err)
	}
}

var usageStr = `
Usage: main.go [options]

Server Options:
    -c, --config <file>              Configuration file path
Common Options:
    -h, --help                       Show this message
    -V, --version                    Show version
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
}
