package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr      string
	Storage       string // "postgres" или "memory"
	DatabaseURL   string
	AdminPassword string
	TokenSecret   string
	TokenTTL      time.Duration
	CORSOrigin    string
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		Storage:       env("STORAGE", "postgres"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		AdminPassword: env("ADMIN_PASSWORD", "admin"),
		TokenSecret:   env("TOKEN_SECRET", "dev-secret-change-me"),
		CORSOrigin:    env("CORS_ORIGIN", "http://localhost:5173"),
	}
	hours, err := strconv.Atoi(env("TOKEN_TTL_HOURS", "12"))
	if err != nil || hours <= 0 {
		return Config{}, errors.New("TOKEN_TTL_HOURS должен быть положительным целым числом")
	}
	c.TokenTTL = time.Duration(hours) * time.Hour

	switch c.Storage {
	case "memory":
	case "postgres":
		if c.DatabaseURL == "" {
			return Config{}, errors.New("для STORAGE=postgres необходимо задать DATABASE_URL")
		}
	default:
		return Config{}, errors.New("STORAGE должен быть «postgres» или «memory»")
	}
	return c, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
