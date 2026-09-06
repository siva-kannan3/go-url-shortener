package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

type URLRepository interface {
	Get(id string) (ShortenedUrl, error)
	Create(shortenedUrl ShortenedUrl) (ShortenedUrl, error)
}

type URLStore struct {
	urls  map[string]ShortenedUrl
	mutex sync.RWMutex
}

func (store *URLStore) Get(id string) (ShortenedUrl, error) {
	store.mutex.RLock()

	defer store.mutex.RUnlock()

	shortenedUrl, exist := store.urls[id]

	if !exist {
		return shortenedUrl, ErrURLNotFound
	}

	return shortenedUrl, nil
}

func (store *URLStore) Create(shortenedURL ShortenedUrl) (ShortenedUrl, error) {

	store.mutex.Lock()

	defer store.mutex.Unlock()

	_, exists := store.urls[shortenedURL.ID]

	if exists {
		return ShortenedUrl{}, ErrIDAlreadyExists
	}

	store.urls[shortenedURL.ID] = shortenedURL

	return shortenedURL, nil
}

func SetupRouter(urlService *URLService) *http.ServeMux {

	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World!"))
	})
	router.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		match, err := urlService.GetUrl(id)
		if errors.Is(err, ErrURLNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectUrl := match.Url
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	})

	router.HandleFunc("POST /shorten", func(w http.ResponseWriter, r *http.Request) {

		requestData := ShortenRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortenedUrl, err := urlService.CreateShortUrl(requestData.Url, requestData.Tags)

		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// 1. Detect the protocol
		scheme := "http"
		// Check if connection is TLS/HTTPS, or if behind a proxy like Nginx/Cloudflare
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		// 2. Build the server URL dynamically using r.Host
		serverURL := fmt.Sprintf("%s://%s", scheme, r.Host)

		// 3. Construct the full URL
		fullURL := fmt.Sprintf("%s/%s", serverURL, shortenedUrl.ID)

		responsePayload := ShortenResponseBody{
			ID:  shortenedUrl.ID,
			Url: fullURL,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responsePayload)
	})
	return router
}
