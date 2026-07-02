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
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /pads/create", app.createPad)
	mux.HandleFunc("POST /pads/create", app.createPadPost)
	mux.HandleFunc("GET /pads/view/{id}", app.viewPad)
	
	return alice.New(app.recoverPanic, app.logRequest, commonHeaders).Then(mux)
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