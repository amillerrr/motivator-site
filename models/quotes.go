package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Quote struct {
	ID       int
	Author   string
	Message  string
	Category string
}

type QuoteService struct {
	DB *pgxpool.Pool
}

// FetchRandomQuote fetches a random quote from the database, optionally filtered by category.
func (service *QuoteService) FetchRandomQuote(category string) (*Quote, error) {
	quote := &Quote{
		Category: category,
	}

	var query string
	if category == "" {
		// Fetch a random quote without any category filter
		query = `SELECT author, message FROM quotes ORDER BY RANDOM() LIMIT 1`
	} else {
		// Fetch a random quote filtered by the given category
		query = `SELECT author, message FROM quotes WHERE category=$1 ORDER BY RANDOM() LIMIT 1`
	}

	// Execute the query
	err := service.DB.QueryRow(context.Background(), query, category).Scan(&quote.Author, &quote.Message)
	if err != nil {
		// Handle case where no rows are found
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no quotes found for category: %s", category)
		}
		return nil, fmt.Errorf("error fetching quote: %v", err)
	}

	return quote, nil
}
