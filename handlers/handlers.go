package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/amillerrr/motivator-site/models"
)

type Template interface {
	Execute(w http.ResponseWriter, r *http.Request, data interface{}) error
}

type Quotes struct {
	Templates struct {
		Quote Template
	}
	QuoteService *models.QuoteService
}

// StaticHandler renders static templates with no dynamic content
func StaticHandler(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tpl.Execute(w, r, nil); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Printf("Error rendering static template: %v", err)
		}
	}
}

// HomePageHandler renders the home page
func (q Quotes) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	err := q.Templates.Quote.Execute(w, r, struct {
		CurrentMessage *models.Quote
		Category       string
	}{
		CurrentMessage: nil, // No initial message
		Category:       "",  // No initial category
	})
	if err != nil {
		http.Error(w, "Error rendering home page", http.StatusInternalServerError)
		log.Printf("Error rendering home page: %v", err)
	}
}

// NewQuoteHandler handles fetching a new quote via HTMX
func (q Quotes) NewQuoteHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	quote, err := q.QuoteService.FetchRandomQuote(category)
	if err != nil {
		http.Error(w, "Error fetching new quote", http.StatusInternalServerError)
		log.Printf("Error fetching quote for category '%s': %v", category, err)
		return
	}

	// Render the quote block for HTMX requests
	err = q.Templates.Quote.Execute(w, r, struct {
		CurrentMessage *models.Quote
		Category       string
	}{
		CurrentMessage: quote,
		Category:       category,
	})
	if err != nil {
		http.Error(w, "Error rendering new quote", http.StatusInternalServerError)
		log.Printf("Error rendering new quote: %v", err)
	}
}

// SetCategoryHandler sets the category via HTMX
func (q Quotes) SetCategoryHandler(w http.ResponseWriter, r *http.Request) {
	category := r.FormValue("category")

	// Validate category input (optional, but useful)
	if category == "" {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		log.Println("Invalid category input")
		return
	}

	// Return the updated hidden input with the selected category
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<input type="hidden" name="category" value="%s">`, category)
}

// GenerateQuoteHandler fetches and renders a quote based on the selected category via HTMX
func (q Quotes) GenerateQuoteHandler(w http.ResponseWriter, r *http.Request) {
	// Get the selected category from the hidden input
	category := r.FormValue("category")

	// Validate category (optional)
	if category == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		log.Println("GenerateQuoteHandler: missing category")
		return
	}

	// Fetch a random quote based on the category
	quote, err := q.QuoteService.FetchRandomQuote(category)
	if err != nil {
		http.Error(w, "Error fetching quote", http.StatusInternalServerError)
		log.Printf("Error fetching quote for category '%s': %v", category, err)
		return
	}

	// Render the quote dynamically for the #quote-container (HTMX request)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
        <blockquote class="text-2xl italic font-semibold text-gray-900">
            "%s"
            <footer class="mt-4 text-gray-500 text-sm">- %s</footer>
        </blockquote>`, quote.Message, quote.Author)
}
