package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

type ShortenRequestBody struct {
	Url  string   `json:"url"`
	Tags []string `json:"tags"`
}

type ShortenResponseBody struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

type ShortenedUrl struct {
	ID   string
	Url  string
	Tags []string
}

func main() {
	// In memory store
	urls := make(map[string]ShortenedUrl)

	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World!"))
	})
	router.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		match, exists := urls[id]
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

		var uniqueId string
		for {
			uniqueGenId := generateId()
			_, exists := urls[uniqueGenId]
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

		urls[uniqueId] = shortenedURL

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

	server := http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	fmt.Println("Listening on PORT: 8000")

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Server is failed to start")
	}
}

func generateRandomCharacter() byte {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	index := rand.Intn(len(characters))

	return characters[index]
}

func generateId() string {

	var id string

	for i := 0; i < 6; i++ {

		id += string(generateRandomCharacter())
	}

	return id
}
