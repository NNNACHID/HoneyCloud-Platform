package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// AttackData represents information captured from web requests
type AttackData struct {
	Timestamp  time.Time         `json:"timestamp"`
	SourceIP   string            `json:"source_ip"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Parameters map[string]string `json:"parameters"`
	AttackType string            `json:"attack_type"`
	Payload    string            `json:"payload"`
}

// Product represents a product in the shop
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// User represents a user account
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

// Comment represents a product review
type Comment struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// PageData holds data for rendering templates
type PageData struct {
	Title       string
	User        *User
	Products    []Product
	Product     *Product
	Comments    []Comment
	Users       []User
	Message     template.HTML
	Error       string
	CurrentYear int
}

var (
	db      *sql.DB
	logFile *os.File
	mutex   sync.Mutex
	tmpls   map[string]*template.Template
)

func main() {
	// Configure logging
	logPath := getEnv("LOG_PATH", "/var/log/honeypot/web_attacks.log")
	logDir := getEnv("LOG_DIR", "/var/log/honeypot")

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Open log file
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Initialize database
	initDB()
	defer db.Close()

	// Initialize templates
	initTemplates()

	// Register HTTP handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/product/", handleProduct)
	http.HandleFunc("/comment", handleComment)
	http.HandleFunc("/admin", handleAdmin)
	http.HandleFunc("/profile/", handleProfile)
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/static/", handleStatic)
	http.HandleFunc("/health", handleHealth)

	// Start Web App server
	port := getEnv("WEB_PORT", "8080")
	log.Printf("Vulnerable Web App Honeypot listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
