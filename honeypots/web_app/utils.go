package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

var (
	// Initialize session store with a secret key
	sessionStore = sessions.NewCookieStore([]byte("vulnerable-honeypot-secret-key"))
)

// getUserFromSession retrieves the current user from session
func getUserFromSession(r *http.Request) *User {
	session, _ := sessionStore.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		return nil
	}

	var user User
	err := db.QueryRow("SELECT id, username, password, email, is_admin FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.IsAdmin)
	if err != nil {
		log.Printf("Error fetching user %d from session: %v", userID, err)
		return nil
	}

	return &user
}

// logAttack records details of attacks to log file
func logAttack(r *http.Request, attackType, payload string) {
	mutex.Lock()
	defer mutex.Unlock()

	// Get client IP
	ip := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = forwardedFor
	}

	// Collect headers
	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[name] = values[0]
	}

	// Collect parameters
	params := make(map[string]string)
	// Query parameters
	for name, values := range r.URL.Query() {
		params[name] = values[0]
	}
	// Form parameters
	r.ParseForm()
	for name, values := range r.Form {
		params[name] = values[0]
	}

	// Create attack data
	attackData := AttackData{
		Timestamp:  time.Now(),
		SourceIP:   ip,
		Method:     r.Method,
		Path:       r.URL.Path,
		Headers:    headers,
		Parameters: params,
		AttackType: attackType,
		Payload:    payload,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(attackData)
	if err != nil {
		log.Printf("Error marshaling attack data: %v", err)
		return
	}

	// Write to log file
	if _, err := logFile.Write(append(jsonData, '\n')); err != nil {
		log.Printf("Error writing to log file: %v", err)
		return
	}

	// Also output to stdout for monitoring
	log.Printf("[ATTACK] %s attack from %s on %s: %s",
		attackType, ip, r.URL.Path, payload)
}

// getEnv returns an environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
