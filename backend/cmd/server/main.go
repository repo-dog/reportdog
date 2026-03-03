package main

import (
	"log"

	"github.com/repo-dog/reportdog/backend/internal/config"
	"github.com/repo-dog/reportdog/backend/internal/db"
	"github.com/repo-dog/reportdog/backend/internal/handlers"
	"github.com/repo-dog/reportdog/backend/internal/router"
	"github.com/repo-dog/reportdog/backend/internal/services"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg)
	db.AutoMigrate(database)

	svc := services.NewReportService(database)
	h := handlers.New(svc)
	r := router.Setup(h, cfg.CORSAllowOrigin, cfg.DisableManualUpload)

	log.Printf("Starting ReportDog server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
