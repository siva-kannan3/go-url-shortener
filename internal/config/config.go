package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	PORT        string
}

func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	port := strings.TrimSpace(os.Getenv("PORT"))

	if databaseURL == "" {
		return Config{}, errors.New("DB url not found")
	}

	if port != "" {
		port = ":8000"
	}

	return Config{
		DatabaseURL: databaseURL,
		PORT:        port,
	}, nil
}
