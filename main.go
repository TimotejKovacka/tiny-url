package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/TimotejKovacka/tiny-url/config"
	"github.com/TimotejKovacka/tiny-url/domains/url"
	"github.com/TimotejKovacka/tiny-url/internal/database"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

const (
	SERVICE_NAME    = "tiny-url-backend"
	SERVICE_VERSION = "1.0.0"
)

func initLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return logger.Sugar(), nil
}

func initObservability(config *config.ConfigYaml, logger *zap.SugaredLogger) (func(), error) {
	ctx := context.Background()
	var OTEL_COLLECTOR = config.Otel.Host + ":" + config.Otel.Port
	logger.Infof("OTEL_COLLECTOR: %s", OTEL_COLLECTOR)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(SERVICE_NAME),
			semconv.ServiceVersionKey.String(SERVICE_VERSION),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(OTEL_COLLECTOR),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Metrics exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(OTEL_COLLECTOR),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(15*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down tracer provider: %v", err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down meter provider: %v", err)
		}
	}

	return cleanup, nil
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

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize observability (tracing, metrics)
	cleanup, err := initObservability(config, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize observability: %v", err)
	}
	defer cleanup()

	// Initialize database
	dbConfig := database.Config{
		Host:     config.DB.Host,
		Port:     config.DB.Port,
		User:     config.DB.Username,
		Password: config.DB.Password,
		DBName:   config.DB.DBName,
	}
	db, err := database.ConnectDB(dbConfig, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	// Create a meter and counter
	meter := otel.Meter("tiny-url-meter")
	urlCounter, _ := meter.Int64Counter("url.created.count")

	// Initialize Gin
	r := gin.New()
	r.Use(otelgin.Middleware(SERVICE_NAME))
	r.Use(gin.Recovery())

	// Add a middleware to inject the logger into the context
	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
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
		BASE_URL: config.Server.Host + ":" + config.Server.Port + "/",
		Logger:   logger,
		Tracer:   otel.Tracer("url-service"),
		URLCounter: func() {
			urlCounter.Add(context.Background(), 1)
		},
	})

	// Start the server
	logger.Info("Starting server on :8080")
	if err := r.Run(":" + config.Server.Port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
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
