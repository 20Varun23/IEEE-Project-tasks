package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
)

// Create a DS for mapping between shortened version and normal URL
var shortURLs = make(map[string]string)

type JSONreq struct {
	URL string `json:"url"`
}

type JSONres struct {
	Short_url string `json:"short_url"`
}

// Generate a random short code
func generateShortCode() string {
	charset := "qwertyuiopasdfghjklzxcvbnm" //defining char set
	str := ""

	for i := 0; i < 5; i++ {
		str = str + string(charset[rand.Intn(26)])
	}

	return str

}

// Shorten URL handler
func shortenURLHandler(w http.ResponseWriter, r *http.Request) {

	// Parse original URL from request body
	reqBody, err1 := io.ReadAll(r.Body)
	if err1 != nil {
		fmt.Println(err1)
	}

	var data JSONreq
	err := json.Unmarshal(reqBody, &data)
	fmt.Printf("%v\n", data)
	if err != nil {
		fmt.Println(err)
	}

	var normalURL = data.URL
	// Generate unique short code
	shortURL := generateShortCode()
	_, exists := shortURLs[shortURL]

	for exists == true {
		shortURL = generateShortCode()
		_, exists = shortURLs[shortURL]
	}

	// Store the mapping
	shortURLs[shortURL] = normalURL
	fmt.Println(normalURL, " ", shortURL)

	// Send back the short URL as response
	resNJ := JSONres{
		Short_url: fmt.Sprintf("http://localhost:8080/r/%s", shortURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resNJ)

}

// Redirect handler
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Path[3:]
	fmt.Println(shortURLs[str])
	http.Redirect(w, r, shortURLs[str], http.StatusFound)
}

// Serve frontend index.html file
func indexHandler(w http.ResponseWriter, r *http.Request) {

	file := "index.html"
	t, err := template.ParseFiles(file)

	if err != nil {
		fmt.Println("Sorry bro, parsing only didnt happen")
	}

	err = t.ExecuteTemplate(w, file, nil)

	if err != nil {
		fmt.Println("Execution failure")
	}
}

func main() {

	// Route for serving the frontend page
	http.HandleFunc("/", indexHandler)

	// Route for the API to shorten URLs
	http.HandleFunc("/shorten", shortenURLHandler)

	// Route for handling redirects
	http.HandleFunc("/r/", redirectHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
