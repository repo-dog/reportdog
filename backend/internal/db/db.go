package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq"

	"github.com/repo-dog/reportdog/backend/internal/config"
)

// Connect establishes a connection to PostgreSQL using database/sql.
func Connect(cfg *config.Config) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("connected to database")
	return db
}

// Migrate reads all .sql files from migrationsDir in alphabetical order
// and executes them against the database.
func Migrate(database *sql.DB, migrationsDir string) {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("failed to read migrations directory: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no migration files found in %s", migrationsDir)
	}

	sort.Strings(files)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("failed to read migration file %s: %v", f, err)
		}
		if _, err := database.Exec(string(content)); err != nil {
			log.Fatalf("migration %s failed: %v", filepath.Base(f), err)
		}
		log.Printf("applied migration: %s", filepath.Base(f))
	}

	log.Println("database migration completed")
}
