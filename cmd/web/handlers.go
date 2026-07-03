package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/discruter/scratchpad/internal/models"
	"github.com/discruter/scratchpad/internal/validator"
)

// Handler function
func (app *application) home(w http.ResponseWriter, r *http.Request) {
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

// struct to hold form data and errors
type padsCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

// Add a pad
func (app *application) createPad(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = padsCreateForm{
		Expires: 365,
	}

	app.render(w, r, http.StatusOK, "create.tmpl", data)
}

// Create a Pad
func (app *application) createPadPost(w http.ResponseWriter, r *http.Request) {
	// PadsCreateForm struct
	var form padsCreateForm
	// Parsing the form data
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// title
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	// content
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	// expires
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// Inserting data
	id, err := app.pads.Insert(form.Title, form.Content, form.Expires)
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
