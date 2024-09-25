package database

import (
	"context"
	"fmt"
	"time"

	"github.com/TimotejKovacka/tiny-url/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
}

type GormLogger struct {
	Logger *zap.SugaredLogger
	Tracer trace.Tracer
}

func NewGormLogger(logger *zap.SugaredLogger) *GormLogger {
	return &GormLogger{
		Logger: logger,
		Tracer: otel.Tracer("gorm-tracer"),
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Infow(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warnw(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Errorw(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	ctx, span := l.Tracer.Start(ctx, "GORM Operation")
	defer span.End()

	span.SetAttributes(
		attribute.String("sql.query", sql),
		attribute.Int64("sql.rows_affected", rows),
		attribute.String("sql.duration", elapsed.String()),
	)

	if err != nil {
		l.Logger.Errorw("GORM query error",
			"error", err,
			"duration", elapsed,
			"rows", rows,
			"sql", sql,
		)
		span.RecordError(err)
	} else {
		l.Logger.Infow("GORM query",
			"duration", elapsed,
			"rows", rows,
			"sql", sql,
		)
	}
}

func ConnectDB(config Config, logger *zap.SugaredLogger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port)

	gormLogger := NewGormLogger(logger)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := registerModels(db); err != nil {
		return nil, fmt.Errorf("failed to register models: %w", err)
	}

	logger.Info("Successfully connected to the database")
	return db, nil
}

func registerModels(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.UrlModel{}); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}
	return nil
}
