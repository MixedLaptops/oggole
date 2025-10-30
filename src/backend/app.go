package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"golang.org/x/crypto/bcrypt"

	_ "modernc.org/sqlite"
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

func main() {
	// Initialize database
	var err error
	db, err = sql.Open("sqlite", "whoknows.db")
	if err != nil {
		panic(err)
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
		rows, err := db.Query("SELECT title, url, language, last_updated, content FROM pages WHERE language = ? AND content LIKE ?", language, "%"+query+"%")
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
	username := r.FormValue("username")
	password := r.FormValue("password")

	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username)
	.Scan(&storedPassword)
	
	// if the result we get back is nil then it worked, if not nill and error was found
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	} else { 
		//Here we check if the storedPassword from db match the one used to login.
		//bcrypt need to convert values to bytes to compare. 
		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
			if err != nil {
				http.Error(w,"Invalid username or password", http.StatusUnauthorized)
				return 
			} else {
				http.SetCookie(w, &http.Cookie{
					Name: "session_token",
					Value: username,
					Path: "/",
				})
				//Here it use w(response writer) to show where to send Redirect
				//r (the request) needed for  context, and "/" the path to be directed to.
				//and http.StatusSeeOther sends a 303 status code, which is like post worked now switch to get and go Here
				http.Redirect(w,r,"/", http.StatusSeeOther)
			}
	}
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
