package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"

	"github.com/amillerrr/motivator-site/handlers"
	"github.com/amillerrr/motivator-site/migrations"
	"github.com/amillerrr/motivator-site/models"
	"github.com/amillerrr/motivator-site/templates"
	"github.com/amillerrr/motivator-site/views"
)

type config struct {
	PSQL   models.PostgresConfig
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, falling back to environment variables.")
	}
	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		Database: os.Getenv("PSQL_DATABASE"),
		SSLMode:  os.Getenv("PSQL_SSLMODE"),
	}
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("missing PSQL configuration (PSQL_HOST or PSQL_PORT)")
	}

	// Load server address (default to :8088 if not set)
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	if cfg.Server.Address == "" {
		cfg.Server.Address = ":8088"
	}

	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = run(cfg)
	if err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run(cfg config) error {
	// Initialize *sql.DB for goose migrations
	sqlDB, err := models.OpenSQLDB(cfg.PSQL)
	if err != nil {
		return fmt.Errorf("Failed to initialize *sql.DB: %w", err)
	}
	defer sqlDB.Close()
	log.Println("Connected to database for goose...")

	if err := runMigrations(sqlDB); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Setup the database connection pool (pgx) for service layer
	log.Println("Setting up PostgreSQL connection pool...")
	dbPool, err := models.OpenPgxPool(cfg.PSQL)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection pool: %w", err)
	}
	defer dbPool.Close()

	// Setup Services and Controllers
	quoteService := &models.QuoteService{DB: dbPool}
	quotesC := handlers.Quotes{
		QuoteService: quoteService,
	}
	quotesC.Templates.Quote = views.Must(views.ParseFS(templates.FS, "quote.gohtml", "tailwind.gohtml"))

	// Setup HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: routes(quotesC),
	}

	// Start the server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("Server starting on %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Graceful shutdown
	return gracefulShutdown(srv)
}

func runMigrations(sqlDB *sql.DB) error {
	log.Println("Running migrations...")
	if err := models.MigrateFS(sqlDB, migrations.FS, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Importing Grit data...")
	if err := migrations.ImportGrit(sqlDB, migrations.DataFS); err != nil {
		return fmt.Errorf("failed to import Grit data: %w", err)
	}

	log.Println("Importing Gratitude data...")
	if err := migrations.ImportGratitude(sqlDB, migrations.DataFS); err != nil {
		return fmt.Errorf("failed to import Gratitude data: %w", err)
	}

	log.Println("Importing Perseverance data...")
	if err := migrations.ImportPerseverance(sqlDB, migrations.DataFS); err != nil {
		return fmt.Errorf("failed to import Perseverance data: %w", err)
	}

	return nil
}

func routes(quotesC handlers.Quotes) http.Handler {
	mux := http.NewServeMux()

	// Define HTTP routes
	mux.HandleFunc("/", quotesC.HomePageHandler)
	mux.HandleFunc("/new-quote", quotesC.NewQuoteHandler)
	mux.HandleFunc("/set-category", quotesC.SetCategoryHandler)
	mux.HandleFunc("/generate-quote", quotesC.GenerateQuoteHandler)

	// Serve static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	return mux
}

func gracefulShutdown(srv *http.Server) error {
	// Channel to listen for interrupt or termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Block until we receive a signal.
	<-stop

	// Create a deadline for the shutdown (5 seconds).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down server...")

	// Attempt a graceful shutdown.
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server shutdown completed.")
	return nil
}
