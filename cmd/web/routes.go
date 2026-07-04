package main

import (
	"net/http"
	"path/filepath"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// HTTP multiplexer
	mux := http.NewServeMux()

	// Fileserve Handle
	fileServer := http.FileServer(nueturedFileSystem{http.Dir("./ui/static/")})
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	// Handler Func Routes
	dynamic := alice.New(app.sessionManager.LoadAndSave) // Dynamic session manager handler
	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /pads/create", dynamic.ThenFunc(app.createPad))
	mux.Handle("POST /pads/create", dynamic.ThenFunc(app.createPadPost))
	mux.Handle("GET /pads/view/{id}", dynamic.ThenFunc(app.viewPad))

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