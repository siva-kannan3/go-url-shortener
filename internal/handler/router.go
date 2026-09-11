package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/siva-kannan3/go-url-shortener/internal/repository"
	"github.com/siva-kannan3/go-url-shortener/internal/service"
)

type ShortenRequestBody struct {
	Url string `json:"url"`
}

type ShortenResponseBody struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

func SetupRouter(urlService *service.URLService) *http.ServeMux {

	router := http.NewServeMux()

	// Health check route
	router.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		match, err := urlService.GetURL(id)
		if errors.Is(err, repository.ErrURLNotFound) {
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

		ShortenedURL, err := urlService.CreateShortUrl(requestData.Url)

		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
		fullURL := fmt.Sprintf("%s/%s", serverURL, ShortenedURL.ID)

		responsePayload := ShortenResponseBody{
			ID:  ShortenedURL.ID,
			Url: fullURL,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responsePayload)
	})
	return router
}
