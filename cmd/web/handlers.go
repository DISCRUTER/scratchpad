package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// Handler function
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "GO")
	// Template files
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/pages/home.tmpl",
		"./ui/html/pages/nav.tmpl",
	}
	// Parsing the template
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send template as response
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Add a pad
func createPad(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new scratchpad..."))
}
// Create a Pad
func createPadPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "GO")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("scratchPad created..."))
}

// View a pad
func viewPad(w http.ResponseWriter, r *http.Request) {
	// Extract & Sanitize `id` value.
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	// Send response
	fmt.Fprintf(w, "Viewing scratchpad %d", id)
}