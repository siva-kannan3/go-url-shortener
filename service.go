package main

import (
	"errors"
	"strings"
)

type URLService struct {
	repository URLRepository
}

func NewURLService(repository URLRepository) *URLService {
	return &URLService{
		repository: repository,
	}
}

func (service *URLService) CreateShortUrl(url string, tags []string) (ShortenedUrl, error) {
	if strings.TrimSpace(url) == "" {
		return ShortenedUrl{}, errors.New("URL is empty")
	}
	return service.repository.Create(url, tags), nil
}

func (service *URLService) GetUrl(id string) (ShortenedUrl, bool) {
	return service.repository.Get(id)
}
