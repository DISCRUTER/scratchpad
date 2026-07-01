package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/discruter/scratchpad/internal/models"
)

type TemplateData struct {
	CurrentYear int
	Pad         models.Pads
	Pads        []models.Pads
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
	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}
	// Iterating through files
	for _, page := range pages {
		// Extracting the base name of the files
		name := filepath.Base(page)

		// Parsing base.tmpl & register template func
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		// Parsing all partial files
		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		// Parsing page
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		// Adding templates to cache map
		cache[name] = ts
	}
	return cache, nil
}
