package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"uptime-up/backend/internal/auth"
	"uptime-up/backend/internal/config"
	"uptime-up/backend/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database connection failed")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("database initialization failed")
	}
	defer sqlDB.Close()
	service := auth.New(db, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	server := &http.Server{Addr: cfg.Addr, Handler: service.Handler(), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("server listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
