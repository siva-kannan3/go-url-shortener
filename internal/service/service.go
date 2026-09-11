package service

import (
	"errors"
	"net/url"
	"strings"

	"github.com/siva-kannan3/go-url-shortener/internal/domain"
	"github.com/siva-kannan3/go-url-shortener/internal/repository"
)

var ErrInvalidURL = errors.New("invalid URL")

type URLService struct {
	repository repository.URLRepository
}

func NewURLService(repository repository.URLRepository) *URLService {
	return &URLService{
		repository: repository,
	}
}

func (service *URLService) CreateShortUrl(url string) (domain.ShortenedURL, error) {
	if !isValidURL(url) {
		return domain.ShortenedURL{}, ErrInvalidURL
	}

	for {
		ShortenedURL := domain.ShortenedURL{
			ID:  generateID(),
			Url: url,
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

func isValidURL(rawURL string) bool {
	trimmedURL := strings.TrimSpace(rawURL)

	if trimmedURL == "" {
		return false
	}

	parsedURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
}
