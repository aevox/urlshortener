package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type RequestBody struct {
	URL string `json:"url"`
}

// URLStore interface for URL storage operations
type URLStore interface {
	InsertURL(slug string, originalURL string) error
	GetOriginalURL(slug string) (string, error)
	FindSlugByURL(originalURL string) (string, error)
}

// PostgresURLStore implements URLStore with PostgreSQL
type PostgresURLStore struct {
	DB *sql.DB
}

func (store *PostgresURLStore) InsertURL(slug string, originalURL string) error {
	_, err := store.DB.Exec("INSERT INTO urls (slug, original_url) VALUES ($1, $2)", slug, originalURL)
	return err
}

func (store *PostgresURLStore) GetOriginalURL(slug string) (string, error) {
	var originalURL string
	err := store.DB.QueryRow("SELECT original_url FROM urls WHERE slug = $1", slug).Scan(&originalURL)
	return originalURL, err
}

func (store *PostgresURLStore) FindSlugByURL(originalURL string) (string, error) {
	var slug string
	err := store.DB.QueryRow("SELECT slug FROM urls WHERE original_url = $1", originalURL).Scan(&slug)
	if err != nil {
		return "", err // Will return sql.ErrNoRows if not found
	}
	return slug, nil
}

func generateSlug(text string) string {
	hash := md5.Sum([]byte(text))
	md5Hex := hex.EncodeToString(hash[:])
	return md5Hex[:6] // Return only the first 6 characters
}

func shortenURL(store URLStore, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("Method is not supported")
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil || requestBody.URL == "" {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the URL is already stored and get the associated slug
	existingSlug, err := store.FindSlugByURL(requestBody.URL)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error querying URL: %v", err)
		http.Error(w, "Error querying URL", http.StatusInternalServerError)
		return
	}

	var slug string
	if err == sql.ErrNoRows {
		// URL not found, generate a new slug and insert it
		slug = generateSlug(requestBody.URL)
		err = store.InsertURL(slug, requestBody.URL)
		if err != nil {
			log.Printf("Error inserting URL: %v", err)
			http.Error(w, "Error storing shortened URL", http.StatusInternalServerError)
			return
		}
	} else {
		// URL found, use the existing slug
		slug = existingSlug
	}

	shortenedURL := "http://localhost:8080/" + slug
	fmt.Printf("Shortened url %s to %s\n", requestBody.URL, shortenedURL)
	json.NewEncoder(w).Encode(map[string]string{"shortened_url": shortenedURL})
}

func redirect(store URLStore, w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, err := store.GetOriginalURL(slug)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	fmt.Printf("Redirected url http://localhost:8080/%s to %s\n", slug, originalURL)
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func initDB() (*sql.DB, error) {
	// Construct the connection string using environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")
	sslmode := "disable" // Typically "require" for production

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

	var db *sql.DB
	var err error

	for i := 0; i < 3; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Error opening database: %v. Retrying in 3 seconds...\n", err)
		} else {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to the database.")
				return db, nil // Return the successful database connection
			}
			log.Printf("Error connecting to the database: %v. Retrying in 3 seconds...\n", err)
		}

		// Wait for 3 seconds before the next retry
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("Error initializing the DB: ", err)
}

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}
	defer db.Close()

	// Create the URLStore implementation
	store := &PostgresURLStore{DB: db}

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		shortenURL(store, w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		redirect(store, w, r)
	})

	fmt.Println("Server started on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
