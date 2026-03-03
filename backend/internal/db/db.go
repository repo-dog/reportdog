package db

import (
	"log"

	"github.com/repo-dog/reportdog/backend/internal/config"
	"github.com/repo-dog/reportdog/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL.
func Connect(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return db
}

// AutoMigrate runs GORM auto-migration for all models.
func AutoMigrate(db *gorm.DB) {
	// Drop legacy normalised tag tables (safe if they don't exist).
	db.Exec("DROP TABLE IF EXISTS report_tags")
	db.Exec("DROP TABLE IF EXISTS tags")

	if err := db.AutoMigrate(
		&models.TestReport{},
		&models.TestSuite{},
		&models.TestCase{},
		&models.KnownTagKey{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// GIN index on the JSONB tags column for fast containment queries.
	db.Exec("CREATE INDEX IF NOT EXISTS idx_test_reports_tags ON test_reports USING GIN (tags)")

	log.Println("database migration completed")
}
