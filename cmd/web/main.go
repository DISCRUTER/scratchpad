package main

import (
	"log"
	"net/http"
)

func main() {
	// HTTP multiplexer
	mux := http.NewServeMux()

	// Fileserve Handle
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	
	// Assigning routes
	// Fileserve routes
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	// Handler Func Routes
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /pad/create", createPad)
	mux.HandleFunc("POST /pad/create", createPadPost)
	mux.HandleFunc("GET /pad/view/{id}", viewPad)
	
	// Starting server...
	log.Print("Starting server on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}