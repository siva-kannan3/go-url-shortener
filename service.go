package main

import (
	"errors"
	"strings"
)

var ErrIDAlreadyExists = errors.New("id already exists")

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

	for {
		shortenedURL := ShortenedUrl{
			ID:   GenerateId(),
			Url:  url,
			Tags: tags,
		}

		result, err := service.repository.Create(shortenedURL)

		if err == nil {
			return result, nil
		}

		if errors.Is(err, ErrIDAlreadyExists) {
			continue
		}

		return ShortenedUrl{}, err
	}

}

func (service *URLService) GetUrl(id string) (ShortenedUrl, bool) {
	return service.repository.Get(id)
}
