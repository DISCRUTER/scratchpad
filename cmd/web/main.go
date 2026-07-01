package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/discruter/scratchpad/internal/models"
	_ "github.com/go-sql-driver/mysql"
)

// Application struct
type application struct {
	logger        *slog.Logger
	pads          *models.PadsModel
	templateCache map[string]*template.Template
}

func main() {
	// Network Address flag
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Data Source Name for MySQL flag
	dsn := flag.String("dsn", "web:pass@tcp(localhost:3306)/scratchpad?parseTime=true", "MySQL data source name.")
	flag.Parse()

	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// DB Conn
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	// Initalizing cache map
	templateCache, err := NewTemplateCahche()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Creating application instance
	app := &application{
		logger:        logger,
		pads:          &models.PadsModel{DB: db},
		templateCache: templateCache,
	}

	// Starting server...
	logger.Info("Staring server", slog.String("addr", *addr))
	err = http.ListenAndServe(*addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	// Start DB connection pooling
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Ping the DB to verify connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
