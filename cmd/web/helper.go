package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		url = r.RequestURI
	)
	app.logger.Error(err.Error(), slog.String("method", method), slog.String("url", url))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data TemplateData) {
	// Retrieve template from cache
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}
	// Creating a buffer
	buf := new(bytes.Buffer)
	// Execute the template inside buffer
	if err := ts.ExecuteTemplate(buf, "base", data); err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Write status code
	w.WriteHeader(status)
	// Writing buffer data to w
	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) TemplateData {
	return TemplateData{
		CurrentYear: time.Now().Year(),
	}
}