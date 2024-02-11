package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockURLStore implements the URLStore interface for testing
type MockURLStore struct {
	// Use a map to simulate the database
	db map[string]string
}

func NewMockURLStore() *MockURLStore {
	return &MockURLStore{db: make(map[string]string)}
}

func (m *MockURLStore) InsertURL(slug string, originalURL string) error {
	m.db[slug] = originalURL
	return nil
}

func (m *MockURLStore) GetOriginalURL(slug string) (string, error) {
	url, exists := m.db[slug]
	if !exists {
		return "", sql.ErrNoRows // Simulate behavior when slug is not found
	}
	return url, nil
}

func TestShortenURL(t *testing.T) {
	store := NewMockURLStore()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortenURL(store, w, r)
	})

	body := &RequestBody{URL: "https://example.com"}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(bodyBytes))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	json.NewDecoder(rr.Body).Decode(&response)

	if response["shortened_url"] == "" {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestRedirect(t *testing.T) {
	store := NewMockURLStore()
	// Insert a URL to test redirection
	store.InsertURL("abc123", "https://example.com")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirect(store, w, r)
	})

	req, _ := http.NewRequest("GET", "/abc123", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	// Check if the Location header is correct
	expectedLocation := "https://example.com"
	location := rr.Header().Get("Location")
	if location != expectedLocation {
		t.Errorf("handler returned wrong Location header: got %v want %v", location, expectedLocation)
	}
}
