package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

// Application struct
type application struct {
	logger *slog.Logger
}

func main() {
	// Network Address flag
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Creating application instance
	app := &application{
		logger: logger,
	}
	
	// Starting server...
	logger.Info("Staring server", slog.String("addr", *addr))
	err := http.ListenAndServe(*addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)
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