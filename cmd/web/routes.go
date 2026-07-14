package main

import (
	"net/http"
	"path/filepath"

	"github.com/discruter/scratchpad/ui"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// HTTP multiplexer
	mux := http.NewServeMux()

	// Test ping
	mux.HandleFunc("GET /ping", ping)

	// Fileserve Handle
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	// Unprotected Routes
	// Consumer Routes
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate) // Dynamic session manager handler
	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /pads/view/{id}", dynamic.ThenFunc(app.viewPad))
	// Auth Routes
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected Routes
	protected := dynamic.Append(app.requireAuthentication)
	// Consumer Routes
	mux.Handle("GET /pads/create", protected.ThenFunc(app.createPad))
	mux.Handle("POST /pads/create", protected.ThenFunc(app.createPadPost))
	// Auth Routes
	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)
}

// Custom FileSystem for FileServer
type nueturedFileSystem struct {
	fs http.FileSystem
}

func (nfs nueturedFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}
	}

	return f, nil
}
