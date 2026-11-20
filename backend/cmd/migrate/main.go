package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
	"uptime-up/backend/internal/database"
)

func run(db *gorm.DB) error {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(756432198)").Error; err != nil {
			return err
		}
		if err := tx.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())").Error; err != nil {
			return err
		}
		for _, name := range names {
			var count int64
			if err := tx.Table("schema_migrations").Where("name = ?", name).Count(&count).Error; err != nil {
				return err
			}
			if count != 0 {
				continue
			}
			sql, err := os.ReadFile(filepath.Join("migrations", name))
			if err != nil {
				return err
			}
			if err := tx.Exec(string(sql)).Error; err != nil {
				return fmt.Errorf("migration %s: %w", name, err)
			}
			if err := tx.Exec("INSERT INTO schema_migrations (name) VALUES (?)", name).Error; err != nil {
				return err
			}
			log.Printf("applied migration %s", name)
		}
		return nil
	})
}
func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := database.Open(dsn)
	if err != nil {
		log.Fatal("database connection failed")
	}
	if err = run(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
