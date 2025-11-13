package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"whoknows/utils"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

// Sætter en general database variable op som kan aktiveres i main
// og sørge for alt der skal tilgå den kan refere til den.
var db *sql.DB
var templates *template.Template

type Page struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Language    string `json:"language"`
	LastUpdated string `json:"last_updated"`
	Content     string `json:"content"`
}

// generateToken creates a cryptographically secure random session token
func generateToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// getCookieSecure determines if cookies should have Secure flag set
// Defaults to true (secure) unless COOKIE_SECURE env var is explicitly "false"
func getCookieSecure() bool {
	cookieSecure := os.Getenv("COOKIE_SECURE")
	return strings.ToLower(cookieSecure) != "false"
}

// createSession stores a new session in the database
func createSession(username, token string) error {
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err := db.Exec(
		"INSERT INTO sessions (token, username, expires_at) VALUES ($1, $2, $3)",
		token, username, expiresAt,
	)
	return err
}

// validateSession checks if a session token is valid and returns the username
func validateSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	var username string
	err = db.QueryRow(
		"SELECT username FROM sessions WHERE token = $1 AND expires_at > $2",
		cookie.Value, time.Now(),
	).Scan(&username)

	return username, err
}

func main() {
	// Check for CLI commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init-db":
			utils.InitDB()
			return
		case "seed-data":
			utils.SeedData()
			return
		}
	}

	// Hent database URL fra environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Initialiser database forbindelse
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Configure connection pooling for production
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	templates = template.Must(template.ParseGlob("templates/*.html"))

	http.HandleFunc("/api/search", search)
	http.HandleFunc("/api/login", login)
	http.HandleFunc("/api/register", register)
	http.HandleFunc("/api/logout", logout)
	http.HandleFunc("/api/weather", weather)
	http.HandleFunc("/login", login1)
	http.HandleFunc("/weather", weather1)
	http.HandleFunc("/register", register1)
	http.HandleFunc("/about", about)
	http.HandleFunc("/", index)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))


	http.ListenAndServe(":8080", nil)
}

func search(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query().Get("q")
	language := request.URL.Query().Get("language")
	if language == "" {
		language = "en"
	}

	var pages []Page

	if query != "" {
		rows, err := db.Query("SELECT title, url, language, last_updated, content FROM pages WHERE language = $1 AND content ILIKE $2", language, "%"+query+"%")
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var page Page
			if err := rows.Scan(&page.Title, &page.URL, &page.Language, &page.LastUpdated, &page.Content); err != nil {
				continue
			}
			pages = append(pages, page)
		}
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(pages)
}

func login(w http.ResponseWriter, r *http.Request){
	// Only accept POST requests, reject GET and other methods
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get client IP for audit logging
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate that both fields have values before processing
	if username == "" || password == "" {
		log.Printf("Login failed: username=%s ip=%s reason=missing_credentials", username, clientIP)
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	// Query for stored password hash
	var storedPassword string
	var userExists bool
	err := db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)

	if err == sql.ErrNoRows {
		// User not found - use dummy hash to prevent timing attacks
		userExists = false
		storedPassword = "$2a$10$N9qo8uLOickgx2ZMRZoMye/XYF4w3KW7QO.hHC5dGxDrKVK5n7C0O" // bcrypt hash of "dummy"
	} else if err != nil {
		// Database error
		log.Printf("Login failed: username=%s ip=%s reason=database_error", username, clientIP)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else {
		userExists = true
	}

	// Always run bcrypt compare to prevent timing attacks
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))

	if err != nil || !userExists {
		// Log the actual reason internally
		if !userExists {
			log.Printf("Login failed: username=%s ip=%s reason=user_not_found", username, clientIP)
		} else {
			log.Printf("Login failed: username=%s ip=%s reason=invalid_password", username, clientIP)
		}
		// Always return the same generic message
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate secure random session token
	token, err := generateToken()
	if err != nil {
		log.Printf("Login failed: username=%s ip=%s reason=token_generation_error", username, clientIP)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Store session in database
	err = createSession(username, token)
	if err != nil {
		log.Printf("Login failed: username=%s ip=%s reason=session_creation_error", username, clientIP)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Successful login
	log.Printf("Login success: username=%s ip=%s", username, clientIP)

	// Set secure session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,                    // Random token
		Path:     "/",
		HttpOnly: true,                     // Prevents JavaScript from accessing the cookie
		Secure:   false,                    // Set true for production HTTPS
		SameSite: http.SameSiteLaxMode,     // CSRF protection
		MaxAge:   86400,                    // Cookie expires after 24 hours (in seconds)
	})

	//Here it use w(response writer) to show where to send Redirect
	//r (the request) needed for  context, and "/" the path to be directed to.
	//and http.StatusSeeOther sends a 303 status code, which is like post worked now switch to get and go Here
	http.Redirect(w,r,"/", http.StatusSeeOther)
}

func register(w http.ResponseWriter, r *http.Request){
	// Only accept POST requests
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get client IP for audit logging
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	// Get form values
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	// Validate all fields are present
	if username == "" || email == "" || password == "" || password2 == "" {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=missing_fields", username, email, clientIP)
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Validate passwords match
	if password != password2 {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=password_mismatch", username, email, clientIP)
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Check if username already exists
	var existingID int
	err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&existingID)
	if err == nil {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=username_taken", username, email, clientIP)
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=database_error", username, email, clientIP)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if email already exists
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&existingID)
	if err == nil {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=email_taken", username, email, clientIP)
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=database_error", username, email, clientIP)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=hash_generation_error", username, email, clientIP)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	// Insert new user
	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		username, email, string(hashedPassword))
	if err != nil {
		log.Printf("Registration failed: username=%s email=%s ip=%s reason=insert_error error=%v", username, email, clientIP, err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	log.Printf("Registration success: username=%s email=%s ip=%s", username, email, clientIP)

	// Auto-login: create session
	token, err := generateToken()
	if err != nil {
		log.Printf("Auto-login failed after registration: username=%s ip=%s reason=token_generation_error", username, clientIP)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = createSession(username, token)
	if err != nil {
		log.Printf("Auto-login failed after registration: username=%s ip=%s reason=session_creation_error", username, clientIP)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	log.Printf("Auto-login success after registration: username=%s ip=%s", username, clientIP)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "logout")
	return
}

func weather(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "weather")
	return
}

func login1(w http.ResponseWriter, r *http.Request){
	templates.ExecuteTemplate(w, "login.html", nil)
}

func weather1(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "weather1")
	return
}

func register1(w http.ResponseWriter, r *http.Request){
	templates.ExecuteTemplate(w, "register.html", nil)
}

func index(w http.ResponseWriter, r *http.Request){
	templates.ExecuteTemplate(w, "search.html", nil)
}

func about(w http.ResponseWriter, r *http.Request){
	templates.ExecuteTemplate(w, "about.html", nil)
}
