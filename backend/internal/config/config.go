package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL, JWTSecret, Addr string
	AccessTTL, RefreshTTL        time.Duration
}

func Load() (Config, error) {
	c := Config{DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SECRET"), Addr: os.Getenv("HTTP_ADDR"), AccessTTL: 15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.JWTSecret) < 32 {
		return c, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	if c.Addr == "" {
		c.Addr = ":8080"
	}
	for _, item := range []struct {
		name string
		dest *time.Duration
	}{{"ACCESS_TTL", &c.AccessTTL}, {"REFRESH_TTL", &c.RefreshTTL}} {
		if value := os.Getenv(item.name); value != "" {
			d, err := time.ParseDuration(value)
			if err != nil || d <= 0 {
				return c, fmt.Errorf("%s must be a positive duration", item.name)
			}
			*item.dest = d
		}
	}
	return c, nil
}
