package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/siva-kannan3/go-url-shortener/internal/repository"
	"github.com/siva-kannan3/go-url-shortener/internal/service"
)

func newTestPostgresService(t *testing.T) (*sql.DB, *service.URLService) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	urlRepository := repository.NewPostgresRepository(db)
	urlService := service.NewURLService(urlRepository)

	return db, urlService
}

func TestShortenURL(t *testing.T) {
	db, urlService := newTestPostgresService(t)

	router := SetupRouter(urlService)

	requestBody := ShortenRequestBody{
		Url: "https://www.google.com",
	}

	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/shorten", &body)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	var response ShortenResponseBody

	err = json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if response.ID == "" {
		t.Fatal("expected ID in response")
	}

	if response.Url == "" {
		t.Fatal("expected Url in response")
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM urls WHERE short_id = $1",
			response.ID,
		)
		if err != nil {
			t.Errorf("failed to clean up test data: %v", err)
		}
	})

	redirectReq := httptest.NewRequest(
		http.MethodGet,
		"/"+response.ID,
		nil,
	)

	redirectRecorder := httptest.NewRecorder()

	router.ServeHTTP(redirectRecorder, redirectReq)

	if redirectRecorder.Code != http.StatusFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusFound,
			redirectRecorder.Code,
		)
	}

	location := redirectRecorder.Header().Get("Location")

	if location != requestBody.Url {
		t.Fatalf(
			"expected redirect location %s, got %s",
			requestBody.Url,
			location,
		)
	}
}

func TestGetShortenURLNotFound(t *testing.T) {
	_, urlService := newTestPostgresService(t)

	router := SetupRouter(urlService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/doesnotexist",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}

func TestShortenURLValidation(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Shorten URL - Invalid JSON",
			body:           `{"url":`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Shorten URL - Empty URL",
			body:           `{"url":""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Shorten URL - Invalid URL",
			body:           `{"url":"not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Shorten URL - Unsupported Scheme",
			body:           `{"url":"ftp://example.com"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, urlService := newTestPostgresService(t)

			router := SetupRouter(urlService)

			requestBody := bytes.NewBufferString(test.body)

			req := httptest.NewRequest(
				http.MethodPost,
				"/shorten",
				requestBody,
			)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					test.expectedStatus,
					recorder.Code,
				)
			}
		})
	}
}
