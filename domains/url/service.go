package url

import (
	"fmt"
	"strings"
)

type URLService struct {
	storage  *URLStorage
	counter  int64
	elements string
}

func NewURLService(storage *URLStorage) *URLService {
	return &URLService{
		storage:  storage,
		counter:  100000000000,
		elements: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}
}

func (s *URLService) LongToShort(url string) (string, error) {
	// Check if URL already exists
	if shortURL, err := s.storage.GetShortURL(url); err == nil {
		return shortURL, nil
	}

	// Generate new short URL
	shortURL := s.base10ToBase62(s.counter)
	s.counter++

	// Store the new URL pair
	err := s.storage.CreateURL(url, shortURL)
	if err != nil {
		return "", err
	}

	return shortURL, nil
}

func (s *URLService) ShortToLong(shortURL string) (string, error) {
	return s.storage.GetLongURL(shortURL)
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
