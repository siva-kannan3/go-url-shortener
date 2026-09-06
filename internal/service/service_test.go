package service

import (
	"errors"
	"testing"

	"github.com/siva-kannan3/go-url-shortener/internal/domain"
	"github.com/siva-kannan3/go-url-shortener/internal/repository"
)

func TestURLServiceCreateShortURLValidation(t *testing.T) {
	mockRepository := MockURLRepository{
		urls: make(map[string]domain.ShortenedURL),
	}

	urlService := NewURLService(&mockRepository)

	_, err := urlService.CreateShortUrl("", []string{})

	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestURLServiceCreateShortURL(t *testing.T) {
	mockRepository := MockURLRepository{
		urls: make(map[string]domain.ShortenedURL),
		createErrors: []error{
			repository.ErrIDAlreadyExists,
			nil,
		},
	}

	urlService := NewURLService(&mockRepository)

	result, err := urlService.CreateShortUrl(
		"https://www.google.com",
		[]string{"google", "website"},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID == "" {
		t.Fatal("expected ID to be generated")
	}

	if result.Url != "https://www.google.com" {
		t.Fatalf(
			"expected URL https://www.google.com, got %s",
			result.Url,
		)
	}

	if mockRepository.createCalls != 2 {
		t.Fatalf("Expected create calls %d, received %d", 2, mockRepository.createCalls)
	}
}

func TestURLServiceGetURL(t *testing.T) {
	mockRepository := MockURLRepository{
		urls: map[string]domain.ShortenedURL{
			"test123": {
				ID:   "test123",
				Url:  "https://www.google.com",
				Tags: []string{"google"},
			},
		},
	}

	urlService := NewURLService(&mockRepository)

	result, err := urlService.GetURL("test123")

	if errors.Is(err, repository.ErrURLNotFound) {
		t.Fatal("expected URL to exist")
	}

	if result.ID != "test123" {
		t.Fatalf("expected ID test123, got %s", result.ID)
	}

	if result.Url != "https://www.google.com" {
		t.Fatalf(
			"expected URL https://www.google.com, got %s",
			result.Url,
		)
	}

	if mockRepository.getCalls != 1 {
		t.Fatalf(
			"expected Get to be called once, got %d",
			mockRepository.getCalls,
		)
	}
}

func TestURLServiceCreateShortURLRepositoryError(t *testing.T) {
	expectedErr := errors.New("database connection failed")

	mockRepository := MockURLRepository{
		urls: make(map[string]domain.ShortenedURL),
		createErrors: []error{
			expectedErr,
		},
	}

	urlService := NewURLService(&mockRepository)

	_, err := urlService.CreateShortUrl(
		"https://www.google.com",
		[]string{"google"},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if mockRepository.createCalls != 1 {
		t.Fatalf(
			"expected create calls %d, received %d",
			1,
			mockRepository.createCalls,
		)
	}
}

type MockURLRepository struct {
	urls         map[string]domain.ShortenedURL
	createCalls  int
	getCalls     int
	createErrors []error
}

func (mock *MockURLRepository) Get(id string) (domain.ShortenedURL, error) {
	mock.getCalls++
	ShortenedURL, exists := mock.urls[id]
	if !exists {
		return domain.ShortenedURL{}, repository.ErrURLNotFound
	}
	return ShortenedURL, nil
}

func (mock *MockURLRepository) Create(ShortenedURL domain.ShortenedURL) (domain.ShortenedURL, error) {
	mock.createCalls++

	if len(mock.createErrors) > 0 {
		err := mock.createErrors[0]
		mock.createErrors = mock.createErrors[1:]

		if err != nil {
			return domain.ShortenedURL{}, err
		}
	}

	mock.urls[ShortenedURL.ID] = ShortenedURL

	return ShortenedURL, nil
}
