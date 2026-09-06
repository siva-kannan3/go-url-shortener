package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
}

func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	if databaseURL == "" {
		return Config{}, errors.New("DB url not found")
	}
	return Config{
		DatabaseURL: databaseURL,
	}, nil
}
