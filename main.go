package main

import (
	"fmt"
	"log"
	"net/http"
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
	urlStore := URLStore{
		urls: make(map[string]ShortenedUrl),
	}

	urlService := NewURLService(&urlStore)

	router := SetupRouter(urlService)

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
