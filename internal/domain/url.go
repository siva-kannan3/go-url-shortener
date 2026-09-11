package domain

import (
	"time"
)

type ShortenedURL struct {
	ID        string
	Url       string
	CreatedAt time.Time
}
