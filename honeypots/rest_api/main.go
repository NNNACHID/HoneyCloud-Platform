package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// AttackData represents information captured from API requests
type AttackData struct {
	Timestamp  time.Time         `json:"timestamp"`
	SourceIP   string            `json:"source_ip"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Parameters map[string]string `json:"parameters"`
	Body       string            `json:"body"`
}

// UserData simulates a user record that might be returned
type UserData struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	APIKey   string `json:"api_key,omitempty"`
}

// fakeDB simulates a database of users
var fakeDB = []UserData{
	{ID: 1, Username: "admin", Email: "admin@example.com", Role: "administrator", APIKey: "12345-secret-api-key"},
	{ID: 2, Username: "user1", Email: "user1@example.com", Role: "user", APIKey: "67890-user1-api-key"},
	{ID: 3, Username: "demo", Email: "demo@example.com", Role: "guest", APIKey: "11111-demo-api-key"},
}

var (
	logFile *os.File
	mutex   sync.Mutex
)

func main() {
	// Configure logging
	logPath := getEnv("LOG_PATH", "/var/log/honeypot/api_requests.log")
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

	// Register HTTP handlers
	http.HandleFunc("/api/users", handleUsers)
	http.HandleFunc("/api/user/", handleUser)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/admin", handleAdmin)

	// Catch-all handler for other endpoints
	http.HandleFunc("/", handleDefault)

	// Start API server
	port := getEnv("API_PORT", "8080")
	log.Printf("REST API Honeypot listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleUsers returns all users in the system (simulated SQL injection vulnerability)
func handleUsers(w http.ResponseWriter, r *http.Request) {
	attackData := captureRequest(r)
	logRequest(attackData)

	query := r.URL.Query().Get("filter")

	// Simulate SQL injection vulnerability
	// If the query contains certain SQL injection patterns, return all user data including API keys
	if query != "" && (contains(query, "OR 1=1") || contains(query, "'") || contains(query, "--")) {
		log.Printf("SQL Injection attempt detected: %s", query)
		// In a real app, this would be where SQL injection would happen
		// We simulate by returning sensitive data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fakeDB)
		return
	}

	// Normal case: return users but strip API keys
	var safeUsers []UserData
	for _, user := range fakeDB {
		// Don't return API keys in normal responses
		user.APIKey = ""
		safeUsers = append(safeUsers, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeUsers)
}

// handleUser returns a specific user by ID (simulates path traversal vulnerability)
func handleUser(w http.ResponseWriter, r *http.Request) {
	attackData := captureRequest(r)
	logRequest(attackData)

	// Extract user ID from path
	userID := r.URL.Path[len("/api/user/"):]

	// Simulate path traversal vulnerability
	if contains(userID, "../") || contains(userID, "etc/passwd") {
		log.Printf("Path traversal attempt detected: %s", userID)
		// Simulate returning the contents of a "sensitive" file
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "root:x:0:0:root:/root:/bin/bash\n")
		fmt.Fprint(w, "daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\n")
		fmt.Fprint(w, "bin:x:2:2:bin:/bin:/usr/sbin/nologin\n")
		return
	}

	// Try to find user by ID
	for _, user := range fakeDB {
		if fmt.Sprintf("%d", user.ID) == userID {
			w.Header().Set("Content-Type", "application/json")
			user.APIKey = "" // Don't return API key
			json.NewEncoder(w).Encode(user)
			return
		}
	}

	// User not found
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error": "User not found"}`)
}

// handleLogin simulates a login endpoint (with authentication bypass vulnerability)
func handleLogin(w http.ResponseWriter, r *http.Request) {
	attackData := captureRequest(r)
	logRequest(attackData)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, `{"error": "Method not allowed"}`)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "Invalid form data"}`)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// Log login attempt
	log.Printf("Login attempt: username=%s, password=%s", username, password)

	// Simulate authentication bypass vulnerability
	// Any username with admin and a specific password pattern will succeed
	if contains(username, "admin") && contains(password, "' OR '1'='1") {
		log.Printf("Authentication bypass attempted and succeeded")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "success", "message": "Login successful", "role": "administrator", "token": "ADMIN_JWT_TOKEN"}`)
		return
	}

	// Hard-coded credentials for demonstration
	if username == "admin" && password == "admin123" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "success", "message": "Login successful", "role": "administrator", "token": "ADMIN_JWT_TOKEN"}`)
		return
	}

	if username == "user1" && password == "password123" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "success", "message": "Login successful", "role": "user", "token": "USER_JWT_TOKEN"}`)
		return
	}

	// Login failed
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "error", "message": "Invalid credentials"}`)
}

// handleAdmin simulates an admin endpoint (with command injection vulnerability)
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	attackData := captureRequest(r)
	logRequest(attackData)

	// Check for admin authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Bearer ADMIN_JWT_TOKEN" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"error": "Unauthorized"}`)
		return
	}

	// Parse the command parameter
	cmd := r.URL.Query().Get("cmd")
	if cmd == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "Missing cmd parameter"}`)
		return
	}

	// Simulate command injection vulnerability
	if contains(cmd, ";") || contains(cmd, "|") || contains(cmd, "`") {
		log.Printf("Command injection attempt detected: %s", cmd)
		// Simulate executing the command and returning results
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Command executed: %s\n", cmd)
		fmt.Fprintf(w, "Output: Simulated command output\n")
		fmt.Fprintf(w, "uid=0(root) gid=0(root) groups=0(root)\n")
		return
	}

	// Normal case
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "success", "command": "%s", "output": "Command executed successfully"}`, cmd)
}

// handleDefault is a catch-all handler
func handleDefault(w http.ResponseWriter, r *http.Request) {
	attackData := captureRequest(r)
	logRequest(attackData)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "error", "message": "Endpoint not found"}`)
}

// captureRequest captures details from an incoming request
func captureRequest(r *http.Request) AttackData {
	// Parse form data if present
	r.ParseForm()

	// Collect headers
	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[name] = values[0]
	}

	// Collect query parameters
	params := make(map[string]string)
	for name, values := range r.URL.Query() {
		params[name] = values[0]
	}

	// Add form parameters
	for name, values := range r.Form {
		params[name] = values[0]
	}

	// Get the source IP
	sourceIP := r.RemoteAddr
	// Check for X-Forwarded-For header
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		sourceIP = forwardedFor
	}

	// Create attack data
	attackData := AttackData{
		Timestamp:  time.Now(),
		SourceIP:   sourceIP,
		Method:     r.Method,
		Path:       r.URL.Path,
		Headers:    headers,
		Parameters: params,
		Body:       "", // We're not capturing the full body for simplicity
	}

	return attackData
}

// logRequest logs the captured request data
func logRequest(data AttackData) {
	mutex.Lock()
	defer mutex.Unlock()

	// Marshal the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling request data: %v", err)
		return
	}

	// Write to log file
	if _, err := logFile.Write(append(jsonData, '\n')); err != nil {
		log.Printf("Error writing to log file: %v", err)
		return
	}

	// Also output to stdout for monitoring
	log.Printf("API Request: %s %s from %s", data.Method, data.Path, data.SourceIP)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s, substr = toLowerCase(s), toLowerCase(substr)
	return s != "" && substr != "" && s != substr && s != strings.Replace(s, substr, "", -1)
}

// toLowerCase converts a string to lowercase
func toLowerCase(s string) string {
	return strings.ToLower(s)
}

// getEnv returns the value of an environment variable or a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
