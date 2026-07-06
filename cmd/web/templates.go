package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/discruter/scratchpad/internal/models"
	"github.com/discruter/scratchpad/ui"
)

type TemplateData struct {
	CurrentYear     int
	Pad             models.Pads
	Pads            []models.Pads
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

// Template functions
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// Creating a template.FuncMap
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func NewTemplateCahche() (map[string]*template.Template, error) {
	// Making cache map
	cache := make(map[string]*template.Template)
	// Getting all the files that match the filepath glob
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}
	// Iterating through files
	for _, page := range pages {
		// Extracting the base name of the files
		name := filepath.Base(page)

		// List all the templates to be parsed
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// Parsing template files
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Adding templates to cache map
		cache[name] = ts
	}
	return cache, nil
}
