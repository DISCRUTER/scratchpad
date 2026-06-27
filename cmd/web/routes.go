package main

import (
	"net/http"
)

func (app *application) routes() *http.ServeMux {
	// HTTP multiplexer
	mux := http.NewServeMux()

	// Fileserve Handle
	fileServer := http.FileServer(nueturedFileSystem{http.Dir("./ui/static/")})
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	// Handler Func Routes
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /pad/create", app.createPad)
	mux.HandleFunc("POST /pad/create", app.createPadPost)
	mux.HandleFunc("GET /pad/view/{id}", app.viewPad)

	return mux
}