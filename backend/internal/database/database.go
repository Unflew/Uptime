package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	// Do not include SQL parameters or connection strings in application logs.
	return gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}
