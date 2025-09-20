package main	

import (
	"database/sql"
	"net/http"
	"fmt"
	"html/template"
	
	_ "modernc.org/sqlite"
)

// Sætter en general database variable op som kan aktiveres i main
// og sørge for alt der skal tilgå den kan refere til den. 
var db *sql.DB
var templates *template.Template 

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

	http.ListenAndServe(":8080", nil)
}

func search(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("search"))	
	return 
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
