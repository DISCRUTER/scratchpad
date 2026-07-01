package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/discruter/scratchpad/internal/models"
)

// Handler function
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "GO")
	// Query database
	pads, err := app.pads.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Constructing TemplateData
	data := app.newTemplateData(r)
	data.Pads = pads

	// Render the page
	app.render(w, r, http.StatusOK, "home.tmpl", data)
}

// Add a pad
func (app *application) createPad(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new scratchpad..."))
}
// Create a Pad
func (app *application) createPadPost(w http.ResponseWriter, r *http.Request) {
	// w.Header().Add("Server", "GO")
	// w.WriteHeader(http.StatusCreated)
	// w.Write([]byte("scratchPad created..."))
	
	// Dummy Data
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7
	// Inserting data
	id, err := app.pads.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, r, err)
		return 
	}
	// Redirect user to view page
	http.Redirect(w, r, fmt.Sprintf("/pads/view/%d", id), http.StatusSeeOther)
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
	// Fetch pad
	pad, err := app.pads.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {	
			app.serverError(w, r, err)
		}
		return
	}

	// Constructing Dynamic data
	data := app.newTemplateData(r)
	data.Pad = pad
	
	// Render the page
	app.render(w, r, http.StatusOK, "view.tmpl", data)
}
