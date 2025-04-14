package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleIndex displays the homepage with all products
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get user from session
	user := getUserFromSession(r)

	// Get all products
	var products []Product
	rows, err := db.Query("SELECT id, name, description, price FROM products")
	if err != nil {
		log.Printf("Error querying products: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price); err != nil {
			log.Printf("Error scanning product row: %v", err)
			continue
		}
		products = append(products, p)
	}

	// Check for reflected XSS vulnerability
	message := r.URL.Query().Get("message")
	if message != "" {
		if strings.Contains(message, "<script>") || 
		   strings.Contains(message, "javascript:") || 
		   strings.Contains(message, "onerror=") {
			logAttack(r, "Reflected XSS", message)
		}
	}

	// Render template with XSS vulnerability (no escaping)
	data := PageData{
		Title:    "Home",
		User:     user,
		Products: products,
		Message:  template.HTML(message), // Intentionally vulnerable to XSS
	}

	tmpl := tmpls["index.html"]
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// handleLogin handles user login with SQL injection vulnerability
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := getUserFromSession(r)
	if user != nil {
		// Already logged in, redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Login form submission
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			renderTemplate := tmpls["login.html"]
			renderTemplate.Execute(w, PageData{
				Title: "Login",
				Error: "Username and password are required",
			})
			return
		}

		// Vulnerable to SQL injection
		// DO NOT USE THIS IN REAL CODE!
		query := fmt.Sprintf("SELECT id, username, email, is_admin FROM users WHERE username = '%s' AND password = '%s'", 
			username, password)

		// Log potential SQL injection attempts
		if strings.Contains(username, "'") || strings.Contains(password, "'") ||
		   strings.Contains(username, "--") || strings.Contains(password, "--") ||
		   strings.Contains(strings.ToLower(username), " or ") || strings.Contains(strings.ToLower(password), " or ") {
			logAttack(r, "SQL Injection", query)
		}

		var u User
		err := db.QueryRow(query).Scan(&u.ID, &u.Username, &u.Email, &u.IsAdmin)
		if err != nil {
			log.Printf("Login failed for user %s: %v", username, err)
			renderTemplate := tmpls["login.html"]
			renderTemplate.Execute(w, PageData{
				Title: "Login",
				Error: "Invalid username or password",
			})
			return
		}

		// Set session
		session, _ := sessionStore.Get(r, "session")
		session.Values["user_id"] = u.ID
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Show login form
	renderTemplate := tmpls["login.html"]
	renderTemplate.Execute(w, PageData{
		Title: "Login",
	})
}

// handleLogout logs out the user
func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleProduct displays a product and its comments (vulnerable to stored XSS)
func handleProduct(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/product/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get user from session
	user := getUserFromSession(r)

	// Get product details
	var product Product
	err = db.QueryRow("SELECT id, name, description, price FROM products WHERE id = ?", id).
		Scan(&product.ID, &product.Name, &product.Description, &product.Price)
	if err != nil {
		log.Printf("Error fetching product %d: %v", id, err)
		http.NotFound(w, r)
		return
	}

	// Get comments for this product
	var comments []Comment
	rows, err := db.Query("SELECT id, product_id, username, content, created_at FROM comments WHERE product_id = ?", id)
	if err != nil {
		log.Printf("Error fetching comments for product %d: %v", id, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var c Comment
			var createdStr string
			if err := rows.Scan(&c.ID, &c.ProductID, &c.Username, &c.Content, &createdStr); err != nil {
				log.Printf("Error scanning comment row: %v", err)
				continue
			}
			c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdStr)
			comments = append(comments, c)
		}
	}

	// Render template
	data := PageData{
		Title:    product.Name,
		User:     user,
		Product:  &product,
		Comments: comments,
	}

	tmpl := tmpls["product.html"]
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// handleComment adds a new comment to a product (vulnerable to stored XSS)
func handleComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get form data
	productID := r.FormValue("product_id")
	content := r.FormValue("content")

	if productID == "" || content == "" {
		http.Error(w, "Product ID and comment are required", http.StatusBadRequest)
		return
	}

	// Get user from session or use "Anonymous"
	user := getUserFromSession(r)
	username := "Anonymous"
	if user != nil {
		username = user.Username
	}

	// Check for potential XSS in comment
	if strings.Contains(content, "<script>") || 
	   strings.Contains(content, "javascript:") || 
	   strings.Contains(content, "onerror=") {
		logAttack(r, "Stored XSS", content)
	}

	// Insert comment without sanitizing content (vulnerable to stored XSS)
	_, err := db.Exec(
		"INSERT INTO comments (product_id, username, content, created_at) VALUES (?, ?, ?, ?)",
		productID, username, content, time.Now().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		http.Error(w, "Error saving comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to product page
	http.Redirect(w, r, "/product/"+productID, http.StatusSeeOther)
}

// handleAdmin displays the admin panel (vulnerable to authorization bypass)
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := getUserFromSession(r)
	if user == nil || !user.IsAdmin {
		// Missing or non-admin user tries to access admin panel
		if user != nil {
			logAttack(r, "Authorization Bypass", fmt.Sprintf("User %s attempted to access admin panel", user.Username))
		} else {
			logAttack(r, "Authorization Bypass", "Unauthenticated user attempted to access admin panel")
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get all users
	var users []User
	rows, err := db.Query("SELECT id, username, email, is_admin FROM users")
	if err != nil {
		log.Printf("Error querying users: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.IsAdmin); err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}
		users = append(users, u)
	}

	// Render admin template
	data := PageData{
		Title: "Admin Panel",
		User:  user,
		Users: users,
	}

	tmpl := tmpls["admin.html"]
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// handleProfile displays a user profile (vulnerable to IDOR)
func handleProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/profile/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get current user from session
	currentUser := getUserFromSession(r)
	if currentUser == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Vulnerable to IDOR - no proper authorization check
	// A user can view any other user's profile including password
	var profileUser User
	err = db.QueryRow("SELECT id, username, password, email, is_admin FROM users WHERE id = ?", id).
		Scan(&profileUser.ID, &profileUser.Username, &profileUser.Password, &profileUser.Email, &profileUser.IsAdmin)
	if err != nil {
		log.Printf("Error fetching user %d: %v", id, err)
		http.NotFound(w, r)
		return
	}

	// Log possible IDOR attack
	if currentUser.ID != profileUser.ID && !currentUser.IsAdmin {
		logAttack(r, "IDOR", fmt.Sprintf("User %s (ID: %d) accessed profile of user %s (ID: %d)",
			currentUser.Username, currentUser.ID, profileUser.Username, profileUser.ID))
	}

	// Render profile template
	data := PageData{
		Title: profileUser.Username + "'s Profile",
		User:  currentUser,
	}
	data.Users = []User{profileUser} // Using Users array to store the profile user

	tmpl := tmpls["profile.html"]
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// handleSearch searches products (vulnerable to SQL injection)
func handleSearch(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := getUserFromSession(r)

	// Get search query
	query := r.URL.Query().Get("q")

	var products []Product
	if query != "" {
		// Vulnerable to SQL injection
		// DO NOT USE THIS IN REAL CODE!
		sqlQuery := fmt.Sprintf("SELECT id, name, description, price FROM products WHERE name LIKE '%%%s%%' OR description LIKE '%%%s%%'",
			query, query)

		// Log potential SQL injection
		if strings.Contains(query, "'") || strings.Contains(query, "--") || strings.Contains(query, ";") {
			logAttack(r, "SQL Injection", sqlQuery)
		}

		rows, err := db.Query(sqlQuery)
		if err != nil {
			log.Printf("Error in search query: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var p Product
				if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price); err != nil {
					log.Printf("Error scanning product row: %v", err)
					continue
				}
				products = append(products, p)
			}
		}
	}

	// Render search template
	data := PageData{
		Title:    "Search Products",
		User:     user,
		Products: products,
		Message:  template.HTML("Search results for: " + query), // Intentionally vulnerable to XSS
	}

	tmpl := tmpls["search.html"]
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// handleStatic serves static files
func handleStatic(w http.ResponseWriter, r *http.Request) {
	// Path traversal vulnerability - no proper path validation
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	
	// Log potential path traversal attacks
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		logAttack(r, "Path Traversal", path)
		
		// Simulate leaking a "sensitive" file for demonstration
		if strings.Contains(path, "passwd") || strings.Contains(path, "shadow") {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "root:x:0:0:root:/root:/bin/bash\n")
			fmt.Fprintf(w, "honeypot:x:1000:1000:Honeypot User:/home/honeypot:/bin/bash\n")
			return
		}
	}
	
	// Actually serve the file (will likely 404 in the honeypot)
	http.ServeFile(w, r, "./static/"+path)
}

// handleHealth is a health check endpoint for Kubernetes
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok"}`)
}
