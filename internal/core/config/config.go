package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/mod-develop/backend/internal/adapters/api/rest"
	"github.com/mod-develop/backend/internal/adapters/storage/database"
)

// Config конфигурация сервиса.
type Config struct {
	LogLevel string `env:"LOG_LEVEL"`
	Rest     rest.Config
	Store    database.Config
}

// Init инициализирует конфигурацию сервиса.
func Init() (*Config, error) {
	cfg := Config{
		Rest: rest.Config{
			Address: "localhost:8080",
		},
	}
	_ = godotenv.Load(".env")

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse config %w", err)
	}

	return &cfg, nil
}
