package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenURL(t *testing.T) {
	router := SetupRouter()

	requestBody := ShortenRequestBody{
		Url:  "https://www.google.com",
		Tags: []string{"google", "website"},
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/shorten", &body)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

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
