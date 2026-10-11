package vote

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port   int    `envconfig:"PORT" default:"59876"`
	DBPath string `envconfig:"DB_PATH" default:"./data/app.db"`
}

var Env Config

// LoadSettings loads a .env file if present, then fills Env from the
// environment. There are no required secrets, so a missing .env is harmless.
func LoadSettings() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using system environment variables instead")
	}

	if err := envconfig.Process("", &Env); err != nil {
		slog.Error("Failed to load environment configuration", "error", err)
		os.Exit(1)
	}

	return &Env
}
