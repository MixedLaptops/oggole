package main	

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type Page struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Language    string `json:"language"`
	LastUpdated string `json:"last_updated"`
	Content     string `json:"content"`
}

func main() {
	http.HandleFunc("/api/search", search)
	http.HandleFunc("/api/login", login)
	http.HandleFunc("/api/register", register)
	http.HandleFunc("/api/logout", logout)
	http.HandleFunc("/api/weather", weather)
	http.HandleFunc("/login", login1)
	http.HandleFunc("/weather", weather1)
	http.HandleFunc("/register", register1)
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
	fmt.Fprint(w, "login1")	
	return
}

func weather1(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "weather1")	
	return
}

func register1(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "register1")	
	return
}

func index(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "index")	
	return
}
