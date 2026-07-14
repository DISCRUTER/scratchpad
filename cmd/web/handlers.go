package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/discruter/scratchpad/internal/models"
	"github.com/discruter/scratchpad/internal/validator"
)

// -------------
// HOME
// -------------

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

// -------------
// PAD CREATION
// -------------

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

	// Add flash message to session data
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	// Redirect user to view page
	http.Redirect(w, r, fmt.Sprintf("/pads/view/%d", id), http.StatusSeeOther)
}

// -------------
// PAD Viewing
// -------------

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

// -------------
// AUTH
// -------------

// Signup Handler
// Signup Form struct
type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.Form = userSignupForm{}

	app.render(w, r, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	// Parsing & Decoding form data
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Data validation
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	// Checking for field errors
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	// Insert user info
	if err := app.users.Insert(form.Name, form.Email, form.Password); err != nil {
		// Check for unique email constraint
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Adding flash message
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Login Handler
// Login form struct
type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validation checks
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	// Check for validation errrors
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}
	// Authenticate user
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
			return
		}
	}
	// Renew Token
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
	}
	// Grant authenticatedUserID
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	// Redirect to pad create
	http.Redirect(w, r, "/pads/create", http.StatusSeeOther)
}

// Logout Handler
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Renew token
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
	}
	// Remove authenticatedUserID
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	// Lodge flash message
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Ping Handler
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
