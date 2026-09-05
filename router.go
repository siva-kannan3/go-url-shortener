package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

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

func (store *URLStore) Create(ctx context.Context, requestData ShortenRequestBody) ShortenedUrl {
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
		Url:  requestData.Url,
		Tags: requestData.Tags,
	}

	store.urls[uniqueId] = shortenedURL

	return shortenedURL
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
		id := r.PathValue("id")
		match, exists := urlStore.Get(id)
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

		if strings.TrimSpace(requestData.Url) == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("URL is empty"))
			return
		}

		shortenedUrl := urlStore.Create(r.Context(), requestData)

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

	router.HandleFunc("GET /context/cancel", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)

		defer cancel()

		err := doSlowOperation123(ctx)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "operation timed out", http.StatusGatewayTimeout)
				return
			}

			if errors.Is(err, context.Canceled) {
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	return router
}

func doSlowOperation123(ctx context.Context) error {
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("work completed")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
