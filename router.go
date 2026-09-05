package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type URLRepository interface {
	Get(id string) (ShortenedUrl, bool)
	Create(url string, tags []string) ShortenedUrl
}

type URLStore struct {
	urls  map[string]ShortenedUrl
	mutex sync.RWMutex
}

func (store *URLStore) Get(id string) (ShortenedUrl, bool) {
	store.mutex.RLock()

	defer store.mutex.RUnlock()

	shortenedUrl, exist := store.urls[id]

	return shortenedUrl, exist
}

func (store *URLStore) Create(url string, tags []string) ShortenedUrl {

	store.mutex.Lock()

	defer store.mutex.Unlock()

	var uniqueId string
	for {
		uniqueGenId := GenerateId()
		_, exists := store.urls[uniqueGenId]
		if !exists {
			uniqueId = uniqueGenId
			break
		}
	}

	shortenedURL := ShortenedUrl{
		ID:   uniqueId,
		Url:  url,
		Tags: tags,
	}

	store.urls[uniqueId] = shortenedURL

	return shortenedURL
}

func SetupRouter(urlService *URLService) *http.ServeMux {

	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World!"))
	})
	router.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		match, exists := urlService.GetUrl(id)
		if !exists {
			http.NotFound(w, r)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
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
