package main	

import (
	"net/http"
)

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
	return 
}

func login(w http.ResponseWriter, r *http.Request){
	return 
}

func register(w http.ResponseWriter, r *http.Request){
 return 
}

func logout(w http.ResponseWriter, r *http.Request){
	return
}

func weather(w http.ResponseWriter, r *http.Request){
	return
}

func login1(w http.ResponseWriter, r *http.Request){
	return
}

func weather1(w http.ResponseWriter, r *http.Request){
	return
}

func register1(w http.ResponseWriter, r *http.Request){
	return
}

func index(w http.ResponseWriter, r *http.Request){
	return
}
