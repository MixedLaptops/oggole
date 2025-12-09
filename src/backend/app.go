package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"whoknows/utils"

	"github.com/joho/godotenv"
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
		rows, err := db.Query("SELECT title, url, language, last_updated, content FROM pages WHERE language = $1 AND content LIKE $2", language, "%"+query+"%")
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
	fmt.Fprint(w, "login")
	return
}

func register(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "login")
	return
}

func logout(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "logout")
	return
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
	// Fetch current weather
	currentURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&units=metric&appid=%s", lat, lon, apiKey)
	currentResp, err := http.Get(currentURL)
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
	forecastResp, err := http.Get(forecastURL)
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
	templates.ExecuteTemplate(w, "login.html", nil)
}

func weather1(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "weather.html", nil)
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