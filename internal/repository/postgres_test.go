package repository

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/siva-kannan3/go-url-shortener/internal/domain"
)

func TestPostgresRepositoryCreateAndGet(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Fatal("DATABASE_URL is not set")
	}
	db, err := sql.Open(
		"pgx",
		databaseURL,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		_, err := db.Exec(
			"DELETE FROM urls WHERE short_id = $1",
			"test123",
		)
		if err != nil {
			t.Errorf("failed to clean up test data: %v", err)
		}
	}()

	repository := NewPostgresRepository(db)

	input := domain.ShortenedURL{
		ID:  "test123",
		Url: "https://example.com",
	}

	_, err = repository.Create(input)
	if err != nil {
		t.Fatal(err)
	}

	result, err := repository.Get("test123")
	if err != nil {
		t.Fatal(err)
	}

	if result.ID != input.ID {
		t.Errorf("expected ID %q, got %q", input.ID, result.ID)
	}

	if result.Url != input.Url {
		t.Errorf("expected URL %q, got %q", input.Url, result.Url)
	}
}

func TestPostgresRepositoryDuplicateID(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Fatal("DATABASE_URL is not set")
	}
	db, err := sql.Open(
		"pgx",
		databaseURL,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	repository := NewPostgresRepository(db)

	input := domain.ShortenedURL{
		ID:  "dupid123",
		Url: "https://example.com",
	}

	defer func() {
		_, err := db.Exec(
			"DELETE FROM urls WHERE short_id = $1",
			input.ID,
		)
		if err != nil {
			t.Errorf("failed to clean up test data: %v", err)
		}
	}()

	_, err = repository.Create(input)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.Create(input)
	if !errors.Is(err, ErrIDAlreadyExists) {
		t.Fatalf("expected ErrIDAlreadyExists, got %v", err)
	}
}

func TestPostgresRepositoryURLNotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Fatal("DATABASE_URL is not set")
	}
	db, err := sql.Open(
		"pgx",
		databaseURL,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	repository := NewPostgresRepository(db)

	input := "abls8232"

	_, err = repository.Get(input)
	if !errors.Is(err, ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound, got %v", err)
	}
}
