package service

import (
	"errors"
	"strings"

	"github.com/siva-kannan3/go-url-shortener/internal/domain"
	"github.com/siva-kannan3/go-url-shortener/internal/repository"
)

type URLService struct {
	repository repository.URLRepository
}

func NewURLService(repository repository.URLRepository) *URLService {
	return &URLService{
		repository: repository,
	}
}

func (service *URLService) CreateShortUrl(url string, tags []string) (domain.ShortenedURL, error) {
	if strings.TrimSpace(url) == "" {
		return domain.ShortenedURL{}, errors.New("URL is empty")
	}

	for {
		ShortenedURL := domain.ShortenedURL{
			ID:   generateID(),
			Url:  url,
			Tags: tags,
		}

		result, err := service.repository.Create(ShortenedURL)

		if err == nil {
			return result, nil
		}

		if errors.Is(err, repository.ErrIDAlreadyExists) {
			continue
		}

		return domain.ShortenedURL{}, err
	}

}

func (service *URLService) GetURL(id string) (domain.ShortenedURL, error) {
	return service.repository.Get(id)
}
