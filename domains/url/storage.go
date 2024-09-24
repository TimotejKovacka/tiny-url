package url

import (
	"errors"

	"github.com/TimotejKovacka/tiny-url/internal/models"
	"gorm.io/gorm"
)

type URLStorage struct {
	db *gorm.DB
}

func NewURLStorage(db *gorm.DB) *URLStorage {
	return &URLStorage{db: db}
}

func (s *URLStorage) CreateURL(longURL, shortURL string) error {
	urlRecord := models.UrlModel{
		LongUrl:  longURL,
		ShortUrl: shortURL,
	}
	return s.db.Create(&urlRecord).Error
}

func (s *URLStorage) GetLongURL(shortURL string) (string, error) {
	var urlRecord models.UrlModel
	result := s.db.Where("short_url = ?", shortURL).First(&urlRecord)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", errors.New("URL not found")
		}
		return "", result.Error
	}
	return urlRecord.LongUrl, nil
}

func (s *URLStorage) GetShortURL(longURL string) (string, error) {
	var urlRecord models.UrlModel
	result := s.db.Where("long_url = ?", longURL).First(&urlRecord)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", errors.New("URL not found")
		}
		return "", result.Error
	}
	return urlRecord.ShortUrl, nil
}

func (s *URLStorage) URLExists(longURL string) (bool, error) {
	var count int64
	result := s.db.Model(&models.UrlModel{}).Where("long_url = ?", longURL).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}
