package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/discruter/scratchpad/internal/models"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// Application struct
type application struct {
	logger         *slog.Logger
	pads           *models.PadsModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
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

	// Intialize form decoder
	formDecoder := form.NewDecoder()

	// Intializing session manager
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// Creating application instance
	app := &application{
		logger:         logger,
		pads:           &models.PadsModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// TLS Config
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	
	// Setting up server
	srv := &http.Server{
		Addr: *addr,
		Handler: app.routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	// Starting server...
	logger.Info("Staring server", slog.String("addr", *addr))
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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
