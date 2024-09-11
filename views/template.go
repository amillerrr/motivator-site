package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"
)

// Public interface for exposing user-friendly error messages.
type public interface {
	Public() string
}

// Must is used to ensure templates are loaded correctly. Panics if there is an error.
func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

// ParseFS parses templates from an embedded filesystem (e.g., using embed.FS).
func ParseFS(fs fs.FS, tplToExecute string, patterns ...string) (Template, error) {
	// Create a new base template
	tpl := template.New(path.Base(tplToExecute))
	tpl = tpl.Funcs(template.FuncMap{
		"errors": func() []string {
			return nil
		},
	})

	// Add tplToExecute to patterns
	patterns = append(patterns, tplToExecute)
	tpl, err := tpl.ParseFS(fs, patterns...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}

	return Template{
		htmlTpl: tpl,
	}, nil
}

// Template represents an HTML template, holding a parsed template.
type Template struct {
	htmlTpl *template.Template
}

// Execute clones the template and renders it to the ResponseWriter.
func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}) error {
	tpl, err := t.htmlTpl.Clone() // Clone template for thread-safety
	if err != nil {
		logTemplateError(w, "cloning template", err)
		return err
	}

	// Set content type before rendering
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create a buffer to capture the output before sending it to the client
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		logTemplateError(w, "executing template", err)
		return err
	}

	// Write the buffer contents to the ResponseWriter
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("error writing response: %v", err)
		http.Error(w, "There was an error writing the response.", http.StatusInternalServerError)
	}
	return nil
}

// logTemplateError is a helper to log and send a standard error response when templates fail.
func logTemplateError(w http.ResponseWriter, action string, err error) {
	log.Printf("%s: %v", action, err)
	http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
}

// errMessages processes multiple errors, returning their user-facing messages.
func errMessages(errs ...error) []string {
	var msgs []string
	for _, err := range errs {
		// Check if the error implements the 'public' interface to expose a user-friendly message
		var pubErr public
		if errors.As(err, &pubErr) {
			msgs = append(msgs, pubErr.Public())
		} else {
			log.Printf("Internal error: %v", err)
			msgs = append(msgs, "Something went wrong.") // Generic message for unexpected errors
		}
	}
	return msgs
}
