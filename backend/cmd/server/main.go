package main

import (
	"log"

	"github.com/akhilbojedla/reportdog/backend/internal/config"
	"github.com/akhilbojedla/reportdog/backend/internal/db"
	"github.com/akhilbojedla/reportdog/backend/internal/handlers"
	"github.com/akhilbojedla/reportdog/backend/internal/router"
	"github.com/akhilbojedla/reportdog/backend/internal/services"
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
