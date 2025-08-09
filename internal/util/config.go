package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

const (
	appConfigFile         = "config.json"
	defaultMigrationsPath = "./internal/db/migrations"
)

type AppConfig struct {
	LogLevel          string        `json:"log_level" validate:"oneof=trace debug info warn error fatal panic"`
	IPBlacklist       []string      `json:"ip_blacklist"` // can contain cidrs
	PubkeyBlacklist   []string      `json:"pubkey_blacklist"`
	CrawlInterval     time.Duration `json:"crawl_interval"`
	DiscoveryTimeout  time.Duration `json:"discovery_timeout"`
	MaxParallelCrawls int           `json:"max_parallel_crawls"`
	MigrationsPath    string        `json:"migrations_path"`
}

type EnvConfig struct {
	DBURL string `env:"DB_URL,notEmpty"`
}

func LoadEnv() *EnvConfig {
	godotenv.Load() // just in case even though env vars should be defined via docker compose
	var cfg EnvConfig
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("Couldn't parse the configuration from environment variables, aborting launch: %v\n", err)
		os.Exit(1)
	}
	return &cfg

}

// Load JSON config to a struct and calculate a SHA-256 hash of the contents.
func LoadConfig() (*AppConfig, [32]byte) {
	content, err := os.ReadFile(appConfigFile)
	if err != nil {
		log.Fatal().Err(err).Str("filepath", appConfigFile).Msg("Error opening the config file")
	}
	sha := sha256.Sum256(content)
	config := &AppConfig{ // apply defaults here
		LogLevel:          "info",
		IPBlacklist:       []string{},
		PubkeyBlacklist:   []string{},
		CrawlInterval:     30 * time.Minute,
		DiscoveryTimeout:  10 * time.Second,
		MaxParallelCrawls: 100,
		MigrationsPath:    defaultMigrationsPath,
	}
	if err := json.Unmarshal(content, &config); err != nil {
		log.Fatal().Err(err).Msg("Error unmarshalling JSON config")
	}
	return config, sha
}
