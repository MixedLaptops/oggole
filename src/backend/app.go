package main	

import (
	"database/sql"
	"net/http"
	"fmt"

	
	_ "modernc.org/sqlite"
)

// Sætter en general database variable op som kan aktiveres i main
// og sørge for alt der skal tilgå den kan refere til den. 
var db *sql.DB

func main() {
	// Initialize database
	var err error
	db, err = sql.Open("sqlite", "whoknows.db")
	if err != nil {
		panic(err)
	}

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
