package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type URLStore struct {
	urls  map[string]ShortenedUrl
	mutex sync.RWMutex
}

func SetupRouter() *http.ServeMux {

	router := http.NewServeMux()
	urlStore := URLStore{
		urls: make(map[string]ShortenedUrl),
	}

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World!"))
	})
	router.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		// urlStore.mutex.RLock()

		// defer urlStore.mutex.RUnlock()

		id := r.PathValue("id")
		match, exists := urlStore.urls[id]
		if !exists {
			http.NotFound(w, r)
			return
		}
		redirectUrl := match.Url
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	})

	router.HandleFunc("POST /shorten", func(w http.ResponseWriter, r *http.Request) {
		urlStore.mutex.Lock()

		defer urlStore.mutex.Unlock()

		requestData := ShortenRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(requestData.Url) == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("URL is empty"))
			return
		}

		var uniqueId string
		for {
			uniqueGenId := GenerateId()
			_, exists := urlStore.urls[uniqueGenId]
			if !exists {
				uniqueId = uniqueGenId
				break
			}
		}

		shortenedURL := ShortenedUrl{
			ID:   uniqueId,
			Url:  requestData.Url,
			Tags: requestData.Tags,
		}

		urlStore.urls[uniqueId] = shortenedURL

		// 1. Detect the protocol
		scheme := "http"
		// Check if connection is TLS/HTTPS, or if behind a proxy like Nginx/Cloudflare
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		// 2. Build the server URL dynamically using r.Host
		serverURL := fmt.Sprintf("%s://%s", scheme, r.Host)

		// 3. Construct the full URL
		fullURL := fmt.Sprintf("%s/%s", serverURL, uniqueId)

		responsePayload := ShortenResponseBody{
			ID:  uniqueId,
			Url: fullURL,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responsePayload)
	})

	return router
}
