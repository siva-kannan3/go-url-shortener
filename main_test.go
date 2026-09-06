package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestShortenURL(t *testing.T) {
	urlStore := URLStore{
		urls: make(map[string]ShortenedUrl),
	}

	urlService := NewURLService(&urlStore)

	router := SetupRouter(urlService)

	requestBody := ShortenRequestBody{
		Url:  "https://www.google.com",
		Tags: []string{"google", "website"},
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
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}

	var response ShortenResponseBody

	err = json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if response.ID == "" {
		t.Fatalf("expected ID in response")
	}

	if response.Url == "" {
		t.Fatalf("expected Url in response")
	}

	redirectReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", response.ID), nil)

	redirectRecorder := httptest.NewRecorder()

	router.ServeHTTP(redirectRecorder, redirectReq)

	if redirectRecorder.Code != http.StatusFound {
		t.Fatalf("Expected status %d, got %d", http.StatusFound, redirectRecorder.Code)
	}

	location := redirectRecorder.Header().Get("Location")
	if location != requestBody.Url {
		t.Fatalf("Expected redirect location %s, got %s", requestBody.Url, location)
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
			expectedStatus: 400,
		},
		{
			name:           "Shorten URL - Empty URL",
			body:           `{"url":""}`,
			expectedStatus: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlStore := URLStore{
				urls: make(map[string]ShortenedUrl),
			}

			urlService := NewURLService(&urlStore)

			router := SetupRouter(urlService)

			requestBody := bytes.NewBufferString(test.body)

			req := httptest.NewRequest(http.MethodPost, "/shorten", requestBody)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != test.expectedStatus {
				t.Fatalf("expected status %d, got %d", test.expectedStatus, recorder.Code)
			}

		})
	}
}

func TestGetShortenURLNotFound(t *testing.T) {
	urlStore := URLStore{
		urls: make(map[string]ShortenedUrl),
	}

	urlService := NewURLService(&urlStore)

	router := SetupRouter(urlService)

	uniqueId := GenerateId()

	req := httptest.NewRequest(http.MethodGet, "/"+uniqueId, nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestConcurrentShortenURL(t *testing.T) {
	urlStore := URLStore{
		urls: make(map[string]ShortenedUrl),
	}

	urlService := NewURLService(&urlStore)

	router := SetupRouter(urlService)
	var wg sync.WaitGroup
	ids := make(chan string, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			requestBody := ShortenRequestBody{
				Url:  "https://www.google.com",
				Tags: []string{"google", "website"},
			}

			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(requestBody)
			if err != nil {
				t.Error(err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/shorten", &body)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusCreated {
				t.Errorf("expected status %d, got %d", http.StatusCreated, recorder.Code)
				return
			}

			var response ShortenResponseBody

			err = json.NewDecoder(recorder.Body).Decode(&response)
			if err != nil {
				t.Error(err)
				return
			}

			if response.ID == "" {
				t.Errorf("expected ID in response")
				return
			}

			if response.Url == "" {
				t.Errorf("expected Url in response")
				return
			}

			ids <- response.ID
		}()
	}

	wg.Wait()

	if len(ids) != 100 {
		t.Errorf("expected 100 IDs, got %d", len(ids))
	}

	uniqueIDs := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := <-ids
		if uniqueIDs[id] {
			t.Fatalf("Duplicate ID already found %s", id)
		}

		uniqueIDs[id] = true
	}
}

func TestURLServiceCreateShortURLValidation(t *testing.T) {
	mockRepository := MockURLRepository{
		urls: make(map[string]ShortenedUrl),
	}

	urlService := NewURLService(&mockRepository)

	_, err := urlService.CreateShortUrl("", []string{})

	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestURLServiceCreateShortURL(t *testing.T) {
	mockRepository := MockURLRepository{
		urls: make(map[string]ShortenedUrl),
	}

	urlService := NewURLService(&mockRepository)

	result, err := urlService.CreateShortUrl(
		"https://www.google.com",
		[]string{"google", "website"},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID == "" {
		t.Fatal("expected ID to be generated")
	}

	if result.Url != "https://www.google.com" {
		t.Fatalf(
			"expected URL https://www.google.com, got %s",
			result.Url,
		)
	}

	if mockRepository.createCalls != 1 {
		t.Fatalf("Expected create calls %d, received %d", 1, mockRepository.createCalls)
	}
}

func TestURLServiceGetURL(t *testing.T) {
	mockRepository := MockURLRepository{
		urls: map[string]ShortenedUrl{
			"test123": {
				ID:   "test123",
				Url:  "https://www.google.com",
				Tags: []string{"google"},
			},
		},
	}

	urlService := NewURLService(&mockRepository)

	result, exists := urlService.GetUrl("test123")

	if !exists {
		t.Fatal("expected URL to exist")
	}

	if result.ID != "test123" {
		t.Fatalf("expected ID test123, got %s", result.ID)
	}

	if result.Url != "https://www.google.com" {
		t.Fatalf(
			"expected URL https://www.google.com, got %s",
			result.Url,
		)
	}

	if mockRepository.getCalls != 1 {
		t.Fatalf(
			"expected Get to be called once, got %d",
			mockRepository.getCalls,
		)
	}
}

type MockURLRepository struct {
	urls        map[string]ShortenedUrl
	createCalls int
	getCalls    int
}

func (mock *MockURLRepository) Get(id string) (ShortenedUrl, bool) {
	mock.getCalls++
	shortenedUrl, exists := mock.urls[id]
	return shortenedUrl, exists
}

func (mock *MockURLRepository) Create(shortenedUrl ShortenedUrl) (ShortenedUrl, error) {
	mock.createCalls++

	mock.urls[shortenedUrl.ID] = shortenedUrl

	return shortenedUrl, nil
}
