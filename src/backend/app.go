package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"whoknows/utils"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// Weather cache structure
type WeatherCache struct {
	Data      WeatherResponse
	ExpiresAt time.Time
	mu        sync.RWMutex
}

// Weather API response structures
type WeatherResponse struct {
	Location    string          `json:"location"`
	Current     CurrentWeather  `json:"current"`
	Forecast    []DailyForecast `json:"forecast"`
	LastUpdated string          `json:"last_updated"`
}

type CurrentWeather struct {
	Temp        float64 `json:"temp"`
	FeelsLike   float64 `json:"feels_like"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
}

type DailyForecast struct {
	Date        string  `json:"date"`
	TempMax     float64 `json:"temp_max"`
	TempMin     float64 `json:"temp_min"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

var weatherCache = &WeatherCache{}

// generateToken creates a cryptographically secure random session token
func generateToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// getClientIP extracts the real client IP, considering proxy headers
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// getCookieSecure determines if cookies should have Secure flag set
// Defaults to true (secure) unless COOKIE_SECURE env var is explicitly "false"
func getCookieSecure() bool {
	cookieSecure := os.Getenv("COOKIE_SECURE")
	return strings.ToLower(cookieSecure) != "false"
}

// setSessionCookie creates and sets a session cookie with consistent security settings
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   getCookieSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
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

// getTextSearchConfig maps and validates language codes to PostgreSQL text search configs
// Whitelisted values prevent SQL injection when building dynamic queries
// NOTE: This mapping must match the CASE statement in init_db.go's update_content_tsv() trigger
func getTextSearchConfig(language string) string {
	switch language {
	case "en":
		return "english"
	case "da":
		return "danish"
	default:
		return "english" // Safe fallback
	}
}

// performSearch executes a search query and returns matching pages
func performSearch(query, language string) ([]Page, error) {
	pages := make([]Page, 0)

	if query == "" {
		return pages, nil
	}

	// Track search query
	searchQueries.Inc()

	// Validate and map language to PostgreSQL text search config
	tsConfig := getTextSearchConfig(language)

	// Hybrid search: full-text search (fast) + ILIKE fallback (partial matching)
	// Note: tsConfig is validated/whitelisted, safe to inject into query string
	searchPattern := "%" + query + "%"
	sqlQuery := fmt.Sprintf(`
		SELECT title, url, language, last_updated, content
		FROM pages
		WHERE language = $1
		  AND (content_tsv @@ plainto_tsquery('%s', $2)
		       OR title ILIKE $3
		       OR content ILIKE $3)
		ORDER BY ts_rank(content_tsv, plainto_tsquery('%s', $2)) DESC
		LIMIT 50
	`, tsConfig, tsConfig)

	rows, err := db.Query(sqlQuery, language, query, searchPattern)

	if err != nil {
		log.Printf("Search query failed: query=%s language=%s error=%v", query, language, err)
		databaseErrors.Inc()
		return pages, err
	}
	defer rows.Close()

	for rows.Next() {
		var page Page
		if err := rows.Scan(&page.Title, &page.URL, &page.Language, &page.LastUpdated, &page.Content); err != nil {
			log.Printf("Search row scan failed: error=%v", err)
			continue
		}
		pages = append(pages, page)
	}

	// Track zero results for quality monitoring
	if len(pages) == 0 {
		searchZeroResults.Inc()
	}

	return pages, nil
}

func main() {
	// Load .env file
	_ = godotenv.Load()

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

	// Verify CRAWLER_API_KEY is set
	if os.Getenv("CRAWLER_API_KEY") == "" {
		log.Fatal("CRAWLER_API_KEY environment variable is required")
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
	http.HandleFunc("/api/batch-pages", batchPages)
	http.HandleFunc("/login", login1)
	http.HandleFunc("/weather", weather1)
	http.HandleFunc("/register", register1)
	http.HandleFunc("/about", about)
	http.HandleFunc("/", index)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Mark service as up
	serviceUp.Set(1)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func search(response http.ResponseWriter, request *http.Request) {
	httpRequestsTotal.WithLabelValues("/search", "200").Inc()

	query := request.URL.Query().Get("q")
	language := request.URL.Query().Get("language")
	if language == "" {
		language = "en"
	}

	pages, err := performSearch(query, language)
	if err != nil {
		httpRequestsTotal.WithLabelValues("/search", "500").Inc()
		http.Error(response, "Search failed", http.StatusInternalServerError)
		return
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
	clientIP := getClientIP(r)

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
		// Log uniform message to prevent username enumeration
		log.Printf("Login failed: ip=%s reason=authentication_failed", clientIP)

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

	// Update user login tracking
	_, err = db.Exec(`UPDATE users
		SET last_login_ip = $1,
			last_login_date = NOW(),
			login_count = login_count + 1
		WHERE username = $2`,
		clientIP, username)
	if err != nil {
		log.Printf("Failed to update login tracking: username=%s error=%v", username, err)
		// Don't fail login if tracking update fails
	}

	// Successful login
	log.Printf("Login success: username=%s ip=%s", username, clientIP)

	// Set secure session cookie
	setSessionCookie(w, token)

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
	clientIP := getClientIP(r)

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
	_, err = db.Exec("INSERT INTO users (username, email, password, registration_ip) VALUES ($1, $2, $3, $4)",
		username, email, string(hashedPassword), clientIP)
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
	setSessionCookie(w, token)

	log.Printf("Auto-login success after registration: username=%s ip=%s", username, clientIP)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request){
	// Get client IP for audit logging
	clientIP := getClientIP(r)

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// No session cookie, redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Delete session from database
	_, err = db.Exec("DELETE FROM sessions WHERE token = $1", cookie.Value)
	if err != nil {
		log.Printf("Logout failed: ip=%s reason=session_deletion_error error=%v", clientIP, err)
	} else {
		log.Printf("Logout success: ip=%s", clientIP)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   getCookieSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Negative value deletes the cookie
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func weather(w http.ResponseWriter, r *http.Request) {
	// Check cache first
	weatherCache.mu.RLock()
	if time.Now().Before(weatherCache.ExpiresAt) && weatherCache.Data.Location != "" {
		data := weatherCache.Data
		weatherCache.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}
	weatherCache.mu.RUnlock()

	// Fetch fresh data if no cache exist
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		http.Error(w, "Weather service not configured", http.StatusServiceUnavailable)
		return
	}

	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")
	if lat == "" || lon == "" {
		lat = "55.6761"  // Copenhagen
		lon = "12.5683"
	}

	weatherData, err := fetchWeatherFromAPI(apiKey, lat, lon)
	if err != nil {
		log.Printf("Failed to fetch weather: %v", err)
		http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
		return
	}

	// Update cache, expires after 15 min
	weatherCache.mu.Lock()
	weatherCache.Data = weatherData
	weatherCache.ExpiresAt = time.Now().Add(15 * time.Minute)
	weatherCache.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weatherData)
}

func fetchWeatherFromAPI(apiKey, lat, lon string) (WeatherResponse, error) {
	// Create HTTP client with timeout to avoid hanging requests
	client := &http.Client{Timeout: 10 * time.Second}

	// Fetch current weather
	currentURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&units=metric&appid=%s", lat, lon, apiKey)
	currentResp, err := client.Get(currentURL)
	if err != nil {
		return WeatherResponse{}, err
	}
	defer currentResp.Body.Close()

	if currentResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(currentResp.Body)
		return WeatherResponse{}, fmt.Errorf("current weather API error: %s - %s", currentResp.Status, string(body))
	}

	var current struct {
		Name  string `json:"name"`
		Coord struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Weather []struct {
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	}

	if err := json.NewDecoder(currentResp.Body).Decode(&current); err != nil {
		return WeatherResponse{}, err
	}

	// Fetch 5 day forecast
	forecastURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%s&lon=%s&units=metric&appid=%s", lat, lon, apiKey)
	forecastResp, err := client.Get(forecastURL)
	if err != nil {
		return WeatherResponse{}, err
	}
	defer forecastResp.Body.Close()

	if forecastResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(forecastResp.Body)
		return WeatherResponse{}, fmt.Errorf("forecast API error: %s - %s", forecastResp.Status, string(body))
	}

	var forecast struct {
		List []struct {
			Dt   int64 `json:"dt"`
			Main struct {
				TempMax float64 `json:"temp_max"`
				TempMin float64 `json:"temp_min"`
			} `json:"main"`
			Weather []struct {
				Description string `json:"description"`
				Icon        string `json:"icon"`
			} `json:"weather"`
		} `json:"list"`
	}

	if err := json.NewDecoder(forecastResp.Body).Decode(&forecast); err != nil {
		return WeatherResponse{}, err
	}

	// Transform to our response format
	location := current.Name
	if location == "" {
		location = fmt.Sprintf("%.2f, %.2f", current.Coord.Lat, current.Coord.Lon)
	}

	weatherResp := WeatherResponse{
		Location:    location,
		LastUpdated: time.Now().Format("2006-01-02 15:04:05"),
		Current: CurrentWeather{
			Temp:      current.Main.Temp,
			FeelsLike: current.Main.FeelsLike,
			Humidity:  current.Main.Humidity,
			WindSpeed: current.Wind.Speed,
		},
	}

	if len(current.Weather) > 0 {
		weatherResp.Current.Description = current.Weather[0].Description
		weatherResp.Current.Icon = current.Weather[0].Icon
	}

	// Process forecast data - group by day and get min/max temps
	dailyForecasts := make(map[string]*DailyForecast)
	var forecastDates []string

	for _, item := range forecast.List {
		dateStr := time.Unix(item.Dt, 0).Format("2006-01-02")
		displayDate := time.Unix(item.Dt, 0).Format("Mon, Jan 2")

		if _, exists := dailyForecasts[dateStr]; !exists {
			forecastDates = append(forecastDates, dateStr)
			dailyForecasts[dateStr] = &DailyForecast{
				Date:    displayDate,
				TempMax: item.Main.TempMax,
				TempMin: item.Main.TempMin,
			}
			if len(item.Weather) > 0 {
				dailyForecasts[dateStr].Description = item.Weather[0].Description
				dailyForecasts[dateStr].Icon = item.Weather[0].Icon
			}
		} else {
			// Update min/max temps
			if item.Main.TempMax > dailyForecasts[dateStr].TempMax {
				dailyForecasts[dateStr].TempMax = item.Main.TempMax
			}
			if item.Main.TempMin < dailyForecasts[dateStr].TempMin {
				dailyForecasts[dateStr].TempMin = item.Main.TempMin
			}
		}
	}

	// Add up to 5 days of forecast (skip today, start from tomorrow)
	for i := 1; i < len(forecastDates) && len(weatherResp.Forecast) < 5; i++ {
		weatherResp.Forecast = append(weatherResp.Forecast, *dailyForecasts[forecastDates[i]])
	}

	return weatherResp, nil
}

func login1(w http.ResponseWriter, r *http.Request){
	if err := templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		log.Printf("Template execution failed: template=login.html error=%v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func weather1(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "weather.html", nil)
}

func register1(w http.ResponseWriter, r *http.Request){
	if err := templates.ExecuteTemplate(w, "register.html", nil); err != nil {
		log.Printf("Template execution failed: template=register.html error=%v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func index(w http.ResponseWriter, r *http.Request){
	query := r.URL.Query().Get("q")
	language := r.URL.Query().Get("language")
	if language == "" {
		language = "en"
	}

	pages, err := performSearch(query, language)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Query":         query,
		"SearchResults": pages,
	}

	if err := templates.ExecuteTemplate(w, "search.html", data); err != nil {
		log.Printf("Template execution failed: template=search.html error=%v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func about(w http.ResponseWriter, r *http.Request){
	if err := templates.ExecuteTemplate(w, "about.html", nil); err != nil {
		log.Printf("Template execution failed: template=about.html error=%v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func batchPages(w http.ResponseWriter, r *http.Request) {
	// Only POST allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check API key
	apiKey := r.Header.Get("X-API-Key")
	expectedKey := os.Getenv("CRAWLER_API_KEY")

	if expectedKey == "" {
		log.Println("WARNING: CRAWLER_API_KEY not set")
		http.Error(w, "Service misconfigured", http.StatusInternalServerError)
		return
	}

	if apiKey != expectedKey {
		log.Printf("Unauthorized crawler request from: %s", getClientIP(r))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse JSON
	var req struct {
		Pages []Page `json:"pages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Pages) == 0 {
		http.Error(w, "No pages provided", http.StatusBadRequest)
		return
	}

	// Insert pages
	success := 0
	errors := 0

	for _, page := range req.Pages {
		if page.Title == "" || page.URL == "" || page.Content == "" {
			errors++
			continue
		}

		if page.Language == "" {
			page.Language = "en"
		}

		_, err := db.Exec(`
			INSERT INTO pages (title, url, language, content, last_updated)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (title)
			DO UPDATE SET url = EXCLUDED.url, content = EXCLUDED.content, last_updated = NOW()
		`, page.Title, page.URL, page.Language, page.Content)

		if err != nil {
			log.Printf("Error inserting page: %v", err)
			errors++
		} else {
			success++
		}
	}

	// Track pages indexed
	pagesIndexed.Add(float64(success))

	// Update total pages count
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM pages").Scan(&count); err == nil {
		totalPages.Set(float64(count))
	}

	log.Printf("Batch insert: success=%d errors=%d total=%d", success, errors, len(req.Pages))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"inserted": success,
		"errors":   errors,
		"total":    len(req.Pages),
	})
}
