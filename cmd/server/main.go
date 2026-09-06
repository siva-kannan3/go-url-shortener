package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/siva-kannan3/go-url-shortener/internal/handler"
	"github.com/siva-kannan3/go-url-shortener/internal/repository"
	"github.com/siva-kannan3/go-url-shortener/internal/service"
)

type ShortenRequestBody struct {
	Url  string   `json:"url"`
	Tags []string `json:"tags"`
}

type ShortenResponseBody struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

func main() {
	urlStore := repository.NewURLStore()

	urlService := service.NewURLService(urlStore)

	router := handler.SetupRouter(urlService)

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
