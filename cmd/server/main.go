package main

import (
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/siva-kannan3/go-url-shortener/internal/config"
	"github.com/siva-kannan3/go-url-shortener/internal/database"
	"github.com/siva-kannan3/go-url-shortener/internal/handler"
	"github.com/siva-kannan3/go-url-shortener/internal/repository"
	"github.com/siva-kannan3/go-url-shortener/internal/service"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgres(cfg.DatabaseURL)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	urlRepository := repository.NewPostgresRepository(db)

	urlService := service.NewURLService(urlRepository)

	router := handler.SetupRouter(urlService)

	server := http.Server{
		Addr:    cfg.PORT,
		Handler: router,
	}

	log.Printf("Server starting on :%s", cfg.PORT)

	log.Fatal(server.ListenAndServe())
}
