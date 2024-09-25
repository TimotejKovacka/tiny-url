package url

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type URLService struct {
	storage  *URLStorage
	counter  int64
	elements string
	tracer   trace.Tracer
	logger   *zap.SugaredLogger
}

func NewURLService(storage *URLStorage, tracer trace.Tracer, logger *zap.SugaredLogger) *URLService {
	return &URLService{
		storage:  storage,
		counter:  100000000000,
		elements: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		tracer:   tracer,
		logger:   logger,
	}
}

func (s *URLService) LongToShort(ctx context.Context, url string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "URLService.LongToShort")
	defer span.End()

	// Check if URL already exists
	if shortURL, err := s.storage.GetShortURL(ctx, url); err == nil {
		s.logger.Infow("URL already exists", "long_url", url, "short_url", shortURL)
		span.SetAttributes(attribute.String("short_url", shortURL))
		span.AddEvent("URL already exists")
		return shortURL, nil
	}

	// Generate new short URL
	shortURL := s.base10ToBase62(s.counter)
	s.counter++

	span.SetAttributes(attribute.String("short_url", shortURL))

	// Store the new URL pair
	err := s.storage.CreateURL(ctx, url, shortURL)
	if err != nil {
		s.logger.Errorw("Failed to create URL", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create URL")
		return "", fmt.Errorf("failed to create URL: %w", err)
	}

	s.logger.Infow("New URL created", "long_url", url, "short_url", shortURL)
	span.AddEvent("New URL created")
	return shortURL, nil
}

func (s *URLService) ShortToLong(ctx context.Context, shortURL string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "URLService.ShortToLong")
	defer span.End()

	span.SetAttributes(attribute.String("short_url", shortURL))

	longURL, err := s.storage.GetLongURL(ctx, shortURL)
	if err != nil {
		s.logger.Errorw("Failed to get long URL", "error", err, "short_url", shortURL)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get long URL")
		return "", fmt.Errorf("failed to get long URL: %w", err)
	}

	s.logger.Infow("Retrieved long URL", "short_url", shortURL, "long_url", longURL)
	span.SetAttributes(attribute.String("long_url", longURL))
	return longURL, nil
}

func (s *URLService) base10ToBase62(n int64) string {
	var sb strings.Builder
	for n != 0 {
		sb.WriteByte(s.elements[n%62])
		n /= 62
	}
	str := sb.String()
	// Reverse the string
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	str = string(runes)
	// Pad with zeros
	return fmt.Sprintf("%07s", str)
}
