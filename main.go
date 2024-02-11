package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// URLStore interface for URL storage operations
type URLStore interface {
	InsertURL(slug string, originalURL string) error
	GetOriginalURL(slug string) (string, error)
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

type RequestBody struct {
	URL string `json:"url"`
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

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %v\n", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Error connecting to the database: %v\n", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS urls (
        slug CHAR(6) PRIMARY KEY,
        original_url TEXT NOT NULL
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("Error creating urls table: %v\n", err)
	}
	fmt.Println("Connected to the database and ensured the 'urls' table exists.")

	return db, nil
}

func generateSlug(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func shortenURL(store URLStore, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil || requestBody.URL == "" {
		log.Printf("%v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slug := generateSlug(6)
	err = store.InsertURL(slug, requestBody.URL)
	if err != nil {
		log.Printf("Error inserting URL: %v", err)
		http.Error(w, "Error storing shortened URL", http.StatusInternalServerError)
		return
	}

	shortenedURL := "http://localhost:8080/" + slug
	json.NewEncoder(w).Encode(map[string]string{"shortened_url": shortenedURL})
}

func redirect(store URLStore, w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, err := store.GetOriginalURL(slug)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Create the URLStore implementation
	store := &PostgresURLStore{DB: db}

	rand.Seed(time.Now().UnixNano())
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
