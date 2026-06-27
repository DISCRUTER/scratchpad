package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// Handler function
func (app *application) home(w http.ResponseWriter, r *http.Request) {
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
		app.serverError(w, r, err)
		return
	}

	// Send template as response
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// Add a pad
func (app *application) createPad(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new scratchpad..."))
}
// Create a Pad
func (app *application) createPadPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "GO")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("scratchPad created..."))
}

// View a pad
func (app *application) viewPad(w http.ResponseWriter, r *http.Request) {
	// Extract & Sanitize `id` value.
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		// http.NotFound(w, r)
		app.clientError(w, http.StatusNotFound)
		return
	}

	// Send response
	fmt.Fprintf(w, "Viewing scratchpad %d", id)
}
