package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

// Sets common headers
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Server", "GO")

		next.ServeHTTP(w, r)
	})
}

// Logs every request
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Declaring info
		var (
			ip = r.RemoteAddr
			proto = r.Proto
			method = r.Method
			uri = r.RequestURI
		)
		// Logging data
		app.logger.Info("received request", slog.String("ip", ip), slog.String("proto", proto), slog.String("method", method), slog.String("uri", uri))

		next.ServeHTTP(w, r)
	})
}

// Recovers from panics
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call a defer function to recover from panic
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}