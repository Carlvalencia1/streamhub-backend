package main

import (
	"log"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/config"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/database"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/database/migrations"
	"github.com/Carlvalencia1/streamhub-backend/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := migrations.Run(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	srv := server.NewServer(cfg, db)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}