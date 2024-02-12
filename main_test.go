package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockURLStore for testing
type MockURLStore struct {
	db map[string]string
}

func NewMockURLStore() *MockURLStore {
	return &MockURLStore{db: make(map[string]string)}
}

// Implement the URLStore interface
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

func (m *MockURLStore) FindSlugByURL(originalURL string) (string, error) {
	for slug, url := range m.db {
		if url == originalURL {
			return slug, nil
		}
	}
	return "", sql.ErrNoRows // Simulate behavior when URL is not found
}

func TestShortenURL(t *testing.T) {
	store := NewMockURLStore()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortenURL(store, w, r)
	})

	// Prepare the request body
	body := RequestBody{URL: "https://example.com"}
	bodyBytes, _ := json.Marshal(body)

	// First call
	req1, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")

	rr1 := httptest.NewRecorder()
	http.HandlerFunc(handler).ServeHTTP(rr1, req1)

	if status := rr1.Code; status != http.StatusOK {
		t.Errorf("First call: handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp1 map[string]string
	err := json.NewDecoder(rr1.Body).Decode(&resp1)
	if err != nil || resp1["shortened_url"] == "" {
		t.Errorf("First call: handler returned unexpected response: %v", err)
	}

	// Second call with the same URL
	req2, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(bodyBytes)) // Create a new request for the second call
	req2.Header.Set("Content-Type", "application/json")

	rr2 := httptest.NewRecorder()
	http.HandlerFunc(handler).ServeHTTP(rr2, req2)

	if status := rr2.Code; status != http.StatusOK {
		t.Errorf("Second call: handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp2 map[string]string
	err = json.NewDecoder(rr2.Body).Decode(&resp2)
	if err != nil || resp2["shortened_url"] == "" {
		t.Errorf("Second call: handler returned unexpected response: %v", err)
	}

	// Compare responses from the first call and second call
	if resp1["shortened_url"] != resp2["shortened_url"] {
		t.Errorf("The shortened URL from the first call (%s) does not match the shortened URL from the second call (%s)", resp1["shortened_url"], resp2["shortened_url"])
	}
}

func TestRedirect(t *testing.T) {
	store := NewMockURLStore()
	// Insert a URL to test redirection
	store.InsertURL("c984d0", "https://example.com")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirect(store, w, r)
	})

	req, _ := http.NewRequest("GET", "/c984d0", nil)
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
