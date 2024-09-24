package urlservice

import (
	"fmt"
	"strings"
)

type URLService struct {
	ltos     map[string]int64
	stol     map[int64]string
	counter  int64
	elements string
}

func NewURLService() *URLService {
	return &URLService{
		ltos:     make(map[string]int64),
		stol:     make(map[int64]string),
		counter:  100000000000,
		elements: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}
}

func (s *URLService) LongToShort(url string) string {
	if existingShort, exists := s.ltos[url]; exists {
		return s.base10ToBase62(existingShort)
	}
	shortURL := s.base10ToBase62(s.counter)
	s.ltos[url] = s.counter
	s.stol[s.counter] = url
	s.counter++
	return shortURL
}

func (s *URLService) ShortToLong(shortURL string) string {
	n := s.base62ToBase10(shortURL)
	return s.stol[n]
}

func (s *URLService) base62ToBase10(str string) int64 {
	var n int64
	for _, c := range str {
		n = n*62 + int64(s.convert(c))
	}
	return n
}

func (s *URLService) convert(c rune) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'z' {
		return int(c-'a') + 10
	}
	if c >= 'A' && c <= 'Z' {
		return int(c-'A') + 36
	}
	return -1
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
