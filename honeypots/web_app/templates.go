package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// initTemplates loads all HTML templates
func initTemplates() {
	tmpls = make(map[string]*template.Template)
	
	// Create templates directory if it doesn't exist
	templateDir := filepath.Join(".", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		log.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Write base layout template
	layoutPath := filepath.Join(templateDir, "layout.html")
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		err := os.WriteFile(layoutPath, []byte(layoutTemplate), 0644)
		if err != nil {
			log.Fatalf("Failed to write layout template: %v", err)
		}
	}
	
	// Write other templates
	templates := map[string]string{
		"index.html":    indexTemplate,
		"login.html":    loginTemplate,
		"product.html":  productTemplate,
		"admin.html":    adminTemplate,
		"profile.html":  profileTemplate,
		"search.html":   searchTemplate,
		"not_found.html": notFoundTemplate,
	}
	
	for name, content := range templates {
		path := filepath.Join(templateDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.WriteFile(path, []byte(content), 0644)
			if err != nil {
				log.Fatalf("Failed to write %s template: %v", name, err)
			}
		}
		
		// Parse template with layout
		tmpl, err := template.New("layout").Parse(layoutTemplate)
		if err != nil {
			log.Fatalf("Failed to parse layout template: %v", err)
		}
		
		tmpl, err = tmpl.Parse(content)
		if err != nil {
			log.Fatalf("Failed to parse %s template: %v", name, err)
		}
		
		tmpls[name] = tmpl
	}
	
	log.Println("Templates loaded successfully")
}

// renderTemplate renders a template with the given data
func renderTemplate(name string, data PageData) template.HTML {
	// Always include current year in page data
	data.CurrentYear = time.Now().Year()
	
	// Render template to string
	tmpl, ok := tmpls[name]
	if !ok {
		log.Printf("Template %s not found", name)
		return template.HTML("Template not found")
	}
	
	// Execute template
	var content strings.Builder
	err := tmpl.Execute(&content, data)
	if err != nil {
		log.Printf("Error executing template %s: %v", name, err)
		return template.HTML("Error rendering template")
	}
	
	return template.HTML(content.String())
}

// Template constants
const layoutTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - TechShop</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 0;
            color: #333;
        }
        header {
            background-color: #2c3e50;
            color: white;
            padding: 1rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        header h1 {
            margin: 0;
        }
        nav ul {
            display: flex;
            list-style: none;
            margin: 0;
            padding: 0;
        }
        nav ul li {
            margin-left: 1rem;
        }
        nav ul li a {
            color: white;
            text-decoration: none;
        }
        nav ul li a:hover {
            text-decoration: underline;
        }
        main {
            max-width: 1200px;
            margin: 0 auto;
            padding: 1rem;
        }
        .container {
            padding: 1rem;
        }
        .alert {
            padding: 1rem;
            margin-bottom: 1rem;
            border-radius: 4px;
        }
        .alert-danger {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .alert-success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .alert-info {
            background-color: #d1ecf1;
            color: #0c5460;
            border: 1px solid #bee5eb;
        }
        .product-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1rem;
        }
        .product-card {
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 1rem;
            transition: transform 0.2s;
        }
        .product-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }
        .btn {
            display: inline-block;
            padding: 0.5rem 1rem;
            background-color: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            border: none;
            cursor: pointer;
            font-size: 1rem;
        }
        .btn:hover {
            background-color: #2980b9;
        }
        .form-group {
            margin-bottom: 1rem;
        }
        .form-group label {
            display: block;
            margin-bottom: 0.5rem;
        }
        .form-control {
            width: 100%;
            padding: 0.5rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 1rem;
        }
        .comment {
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 1rem;
            margin-bottom: 1rem;
        }
        footer {
            background-color: #2c3e50;
            color: white;
            text-align: center;
            padding: 1rem;
            margin-top: 2rem;
        }
    </style>
</head>
<body>
    <header>
        <h1>TechShop</h1>
        <nav>
            <ul>
                <li><a href="/">Products</a></li>
                <li><a href="/search">Search</a></li>
                {{if .User}}
                    {{if .User.IsAdmin}}
                        <li><a href="/admin">Admin Panel</a></li>
                    {{end}}
                    <li><a href="/profile/{{.User.ID}}">Profile</a></li>
                    <li><a href="/logout">Logout</a></li>
                {{else}}
                    <li><a href="/login">Login</a></li>
                {{end}}
            </ul>
        </nav>
    </header>
    
    <main>
        {{template "content" .}}
    </main>
    
    <footer>
        <p>&copy; {{.CurrentYear}} TechShop - This is a honeypot application</p>
    </footer>
</body>
</html>`

const indexTemplate = `{{define "content"}}
<div class="container">
    <h1>Welcome to TechShop</h1>
    
    {{if .Message}}
        <div class="alert alert-info">
            {{.Message}}
        </div>
    {{end}}
    
    <p>Discover our latest technology products with the best prices!</p>
    
    <div class="product-grid">
        {{range .Products}}
            <div class="product-card">
                <h3>{{.Name}}</h3>
                <p>{{.Description}}</p>
                <p><strong>${{.Price}}</strong></p>
                <a href="/product/{{.ID}}" class="btn">View Details</a>
            </div>
        {{end}}
    </div>
</div>
{{end}}`

const loginTemplate = `{{define "content"}}
<div class="container">
    <h1>Login</h1>
    
    {{if .Error}}
        <div class="alert alert-danger">
            {{.Error}}
        </div>
    {{end}}
    
    <form action="/login" method="POST">
        <div class="form-group">
            <label for="username">Username</label>
            <input type="text" class="form-control" id="username" name="username" required>
            <small>Try: admin, user1, or user2</small>
        </div>
        
        <div class="form-group">
            <label for="password">Password</label>
            <input type="password" class="form-control" id="password" name="password" required>
            <small>Try: admin123, password123, or letmein</small>
        </div>
        
        <button type="submit" class="btn">Login</button>
    </form>
    
    <p><small>Hint: This form is vulnerable to SQL injection</small></p>
</div>
{{end}}`

const productTemplate = `{{define "content"}}
<div class="container">
    <h1>{{.Product.Name}}</h1>
    <p>{{.Product.Description}}</p>
    <p><strong>${{.Product.Price}}</strong></p>
    
    <hr>
    
    <h2>Customer Reviews</h2>
    
    {{if .Comments}}
        {{range .Comments}}
            <div class="comment">
                <h4>{{.Username}}</h4>
                <p>{{.Content}}</p>
                <small>Posted on {{.CreatedAt.Format "Jan 02, 2006"}}</small>
            </div>
        {{end}}
    {{else}}
        <p>No reviews yet. Be the first to comment!</p>
    {{end}}
    
    <h3>Add a Review</h3>
    <form action="/comment" method="POST">
        <input type="hidden" name="product_id" value="{{.Product.ID}}">
        
        <div class="form-group">
            <label for="content">Your Review</label>
            <textarea class="form-control" id="content" name="content" rows="3" required></textarea>
            <small>This form is vulnerable to stored XSS attacks.</small>
        </div>
        
        <button type="submit" class="btn">Submit Review</button>
    </form>
    
    <p><a href="/" class="btn">Back to Products</a></p>
</div>
{{end}}`

const adminTemplate = `{{define "content"}}
<div class="container">
    <h1>Admin Panel</h1>
    
    <h2>Users</h2>
    <table style="width:100%; border-collapse: collapse;">
        <thead>
            <tr style="background-color: #f2f2f2;">
                <th style="padding: 8px; border: 1px solid #ddd;">ID</th>
                <th style="padding: 8px; border: 1px solid #ddd;">Username</th>
                <th style="padding: 8px; border: 1px solid #ddd;">Email</th>
                <th style="padding: 8px; border: 1px solid #ddd;">Admin</th>
                <th style="padding: 8px; border: 1px solid #ddd;">Profile</th>
            </tr>
        </thead>
        <tbody>
            {{range .Users}}
                <tr>
                    <td style="padding: 8px; border: 1px solid #ddd;">{{.ID}}</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">{{.Username}}</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">{{.Email}}</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">{{if .IsAdmin}}Yes{{else}}No{{end}}</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">
                        <a href="/profile/{{.ID}}" class="btn">View Profile</a>
                    </td>
                </tr>
            {{end}}
        </tbody>
    </table>
</div>
{{end}}`

const profileTemplate = `{{define "content"}}
<div class="container">
    <h1>User Profile</h1>
    
    <div style="border: 1px solid #ddd; border-radius: 4px; padding: 1rem;">
        <h2>{{.User.Username}}</h2>
        <p><strong>Email:</strong> {{.User.Email}}</p>
        <p><strong>Admin:</strong> {{if .User.IsAdmin}}Yes{{else}}No{{end}}</p>
        <p><strong>Password:</strong> {{.User.Password}}</p>
    </div>
    
    <p><a href="/" class="btn">Back to Products</a></p>
</div>
{{end}}`

const searchTemplate = `{{define "content"}}
<div class="container">
    <h1>Search Products</h1>
    
    <form action="/search" method="GET">
        <div class="form-group">
            <input type="text" class="form-control" name="q" placeholder="Search for products..." value="{{.Query}}">
        </div>
        <button type="submit" class="btn">Search</button>
    </form>
    
    {{if .Message}}
        <div class="alert alert-info">
            {{.Message}}
        </div>
    {{end}}
    
    {{if .Products}}
        <h2>Search Results</h2>
        <div class="product-grid">
            {{range .Products}}
                <div class="product-card">
                    <h3>{{.Name}}</h3>
                    <p>{{.Description}}</p>
                    <p><strong>${{.Price}}</strong></p>
                    <a href="/product/{{.ID}}" class="btn">View Details</a>
                </div>
            {{end}}
        </div>
    {{end}}
</div>
{{end}}`

const notFoundTemplate = `{{define "content"}}
<div class="container">
    <h1>404 - Page Not Found</h1>
    <p>The page you are looking for does not exist.</p>
    <p><a href="/" class="btn">Back to Home</a></p>
</div>
{{end}}`
