package repository

import (
	"errors"
	"sync"

	"github.com/siva-kannan3/go-url-shortener/internal/domain"
)

var (
	ErrIDAlreadyExists = errors.New("id already exists")
	ErrURLNotFound     = errors.New("url not found")
)

type URLRepository interface {
	Get(id string) (domain.ShortenedURL, error)
	Create(ShortenedURL domain.ShortenedURL) (domain.ShortenedURL, error)
}

type URLStore struct {
	urls  map[string]domain.ShortenedURL
	mutex sync.RWMutex
}

func NewURLStore() *URLStore {
	return &URLStore{
		urls: make(map[string]domain.ShortenedURL),
	}
}

func (store *URLStore) Get(id string) (domain.ShortenedURL, error) {
	store.mutex.RLock()

	defer store.mutex.RUnlock()

	ShortenedURL, exist := store.urls[id]

	if !exist {
		return ShortenedURL, ErrURLNotFound
	}

	return ShortenedURL, nil
}

func (store *URLStore) Create(ShortenedURL domain.ShortenedURL) (domain.ShortenedURL, error) {

	store.mutex.Lock()

	defer store.mutex.Unlock()

	_, exists := store.urls[ShortenedURL.ID]

	if exists {
		return domain.ShortenedURL{}, ErrIDAlreadyExists
	}

	store.urls[ShortenedURL.ID] = ShortenedURL

	return ShortenedURL, nil
}
