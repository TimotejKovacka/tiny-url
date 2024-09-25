package url

import (
	"context"
	"errors"
	"fmt"

	"github.com/TimotejKovacka/tiny-url/internal/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type URLStorage struct {
	db     *gorm.DB
	tracer trace.Tracer
	logger *zap.SugaredLogger
}

func NewURLStorage(db *gorm.DB, tracer trace.Tracer, logger *zap.SugaredLogger) *URLStorage {
	return &URLStorage{db: db, tracer: tracer, logger: logger}
}

func (s *URLStorage) CreateURL(ctx context.Context, longURL, shortURL string) error {
	ctx, span := s.tracer.Start(ctx, "URLStorage.CreateURL")
	defer span.End()

	span.SetAttributes(
		attribute.String("long_url", longURL),
		attribute.String("short_url", shortURL),
	)

	urlRecord := models.UrlModel{
		LongUrl:  longURL,
		ShortUrl: shortURL,
	}
	err := s.db.WithContext(ctx).Create(&urlRecord).Error
	if err != nil {
		s.logger.Errorw("Failed to create URL record", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create URL record")
		return fmt.Errorf("failed to create URL record: %w", err)
	}
	return nil
}

func (s *URLStorage) GetLongURL(ctx context.Context, shortURL string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "URLStorage.GetLongURL")
	defer span.End()

	span.SetAttributes(attribute.String("short_url", shortURL))

	var urlRecord models.UrlModel
	result := s.db.WithContext(ctx).Where("short_url = ?", shortURL).First(&urlRecord)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warnw("URL not found", "short_url", shortURL)
			span.SetStatus(codes.Error, "URL not found")
			return "", errors.New("URL not found")
		}
		s.logger.Errorw("Database error", "error", result.Error)
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "Database error")
		return "", fmt.Errorf("database error: %w", result.Error)
	}
	span.SetAttributes(attribute.String("long_url", urlRecord.LongUrl))
	return urlRecord.LongUrl, nil
}

func (s *URLStorage) GetShortURL(ctx context.Context, longURL string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "URLStorage.GetShortURL")
	defer span.End()

	span.SetAttributes(attribute.String("long_url", longURL))

	var urlRecord models.UrlModel
	result := s.db.WithContext(ctx).Where("long_url = ?", longURL).First(&urlRecord)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warnw("URL not found", "long_url", longURL)
			span.SetStatus(codes.Error, "URL not found")
			return "", errors.New("URL not found")
		}
		s.logger.Errorw("Database error", "error", result.Error)
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "Database error")
		return "", fmt.Errorf("database error: %w", result.Error)
	}
	span.SetAttributes(attribute.String("short_url", urlRecord.ShortUrl))
	return urlRecord.ShortUrl, nil
}

func (s *URLStorage) URLExists(ctx context.Context, longURL string) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "URLStorage.URLExists")
	defer span.End()

	span.SetAttributes(attribute.String("long_url", longURL))

	var count int64
	result := s.db.WithContext(ctx).Model(&models.UrlModel{}).Where("long_url = ?", longURL).Count(&count)
	if result.Error != nil {
		s.logger.Errorw("Database error", "error", result.Error)
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "Database error")
		return false, fmt.Errorf("database error: %w", result.Error)
	}
	exists := count > 0
	span.SetAttributes(attribute.Bool("exists", exists))
	return exists, nil
}
