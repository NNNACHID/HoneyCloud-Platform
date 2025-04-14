package main

import (
	"database/sql"
	"log"
)

// initDB initializes the SQLite database and sample data
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create users table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		email TEXT,
		is_admin INTEGER DEFAULT 0
	)`)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// Create products table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Failed to create products table: %v", err)
	}

	// Create comments table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER,
		username TEXT,
		content TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Failed to create comments table: %v", err)
	}

	// Insert sample users if none exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query users count: %v", err)
	}

	if count == 0 {
		// Insert sample users
		_, err = db.Exec(`INSERT INTO users (username, password, email, is_admin) VALUES 
			('admin', 'admin123', 'admin@example.com', 1),
			('user1', 'password123', 'user1@example.com', 0),
			('user2', 'letmein', 'user2@example.com', 0)`)
		if err != nil {
			log.Fatalf("Failed to insert sample users: %v", err)
		}

		// Insert sample products
		_, err = db.Exec(`INSERT INTO products (name, description, price) VALUES 
			('Laptop', 'High performance laptop', 999.99),
			('Smartphone', '5G Smartphone with great camera', 699.99),
			('Headphones', 'Noise cancelling headphones', 199.99)`)
		if err != nil {
			log.Fatalf("Failed to insert sample products: %v", err)
		}

		// Insert sample comments
		_, err = db.Exec(`INSERT INTO comments (product_id, username, content, created_at) VALUES 
			(1, 'user1', 'Great laptop, very fast!', '2023-04-01 10:30:00'),
			(2, 'user2', 'Love the camera on this phone', '2023-04-02 15:45:00')`)
		if err != nil {
			log.Fatalf("Failed to insert sample comments: %v", err)
		}

		log.Println("Sample data inserted successfully")
	}
}
