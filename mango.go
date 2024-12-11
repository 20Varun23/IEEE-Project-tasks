package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"sync"
)

// Create a DS for mapping between shortened version and normal URL

type SafeMap struct {
	mu        sync.Mutex
	shortURLs map[string]string
}

type JSONreq struct {
	URL string `json:"url"`
}

type JSONres struct {
	Short_url string `json:"short_url"`
}

// Generate a random short code
func generateShortCode() string {
	charset := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890" //defining char set
	str := ""

	for i := 0; i < 5; i++ {
		str = str + string(charset[rand.Intn(26)])
	}

	return str

}

// Shorten URL handler
func (s *SafeMap) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	_, exists := s.shortURLs[shortURL]

	for exists == true {
		shortURL = generateShortCode()
		_, exists = s.shortURLs[shortURL]
	}

	// Store the mapping
	s.shortURLs[shortURL] = normalURL
	fmt.Println(normalURL, " ", shortURL)

	// Send back the short URL as response
	resNJ := JSONres{
		Short_url: fmt.Sprintf("http://localhost:8080/r/%s", shortURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resNJ)

}

// Redirect handler
func (s *SafeMap) redirectHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	str := r.URL.Path[3:]
	fmt.Println(s.shortURLs[str])
	http.Redirect(w, r, s.shortURLs[str], http.StatusFound)
}

// Serve frontend index.html file
func indexHandler(w http.ResponseWriter, r *http.Request) {

	file := "index.html"
	t, err := template.ParseFiles(file)

	if err != nil {
		fmt.Println("Parsing failed")
	}

	err = t.ExecuteTemplate(w, file, nil)

	if err != nil {
		fmt.Println("Execution failure")
	}
}

func main() {

	var s = SafeMap{shortURLs: make(map[string]string)}

	// Route for serving the frontend page
	http.HandleFunc("/", indexHandler)

	// Route for the API to shorten URLs
	http.HandleFunc("/shorten", s.shortenURLHandler)

	// Route for handling redirects
	http.HandleFunc("/r/", s.redirectHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
