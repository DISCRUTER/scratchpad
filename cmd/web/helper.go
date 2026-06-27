package main

import (
	"net/http"
	"log/slog"
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