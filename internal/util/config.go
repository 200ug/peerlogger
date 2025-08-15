package util

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type EnvConfig struct {
	LogLevel            string `env:"LOG_LEVEL" envDefault:"info"`
	DBURL               string `env:"DB_URL,notEmpty"`
	IPBlacklistPath     string `env:"IP_BLACKLIST_PATH"`
	PubkeyBlacklistPath string `env:"PUBKEY_BLACKLIST_PATH"`
	GeoIPCityDBPath     string `env:"GEOIP_CITY_DB_PATH,notEmpty"`
	GeoIPASNDBPath      string `env:"GEOIP_ASN_DB_PATH,notEmpty"`
	ENRDBPath           string `env:"ENR_DB_PATH" envDefault:"./enr-data/enode.db"`
}

func LoadEnv() *EnvConfig {
	// load .env just in case (even though envs should be defined in the compose file)
	if err := godotenv.Load(); err != nil {
		// not an actual error in prod. (see above)
		log.Debug().Msg("File .env not found, using system environment variables")
	}
	var cfg EnvConfig
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("Couldn't parse the configuration from environment variables, aborting launch: %v\n", err)
		os.Exit(1)
	}
	return &cfg
}

func LoadJSONList(path, key string) []string {
	if path == "" {
		log.Debug().Str("path", path).Str("key", key).Msg("Empty path provided, returning empty list")
		return []string{}
	}

	// Read the JSON file
	data, err := os.ReadFile(path)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Str("key", key).Msg("Failed to read JSON file, returning empty list")
		return []string{}
	}

	// Parse the JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Warn().Err(err).Str("path", path).Str("key", key).Msg("Failed to parse JSON file, returning empty list")
		return []string{}
	}

	// Extract the specified key
	value, exists := jsonData[key]
	if !exists {
		log.Warn().Str("path", path).Str("key", key).Msg("Key not found in JSON file, returning empty list")
		return []string{}
	}

	// Convert to []string
	switch v := value.(type) {
	case []interface{}:
		list := make([]string, 0, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				list = append(list, str)
			} else {
				log.Warn().Str("path", path).Str("key", key).Int("index", i).
					Msg("Non-string value found in array, skipping")
			}
		}
		log.Debug().Str("path", path).Str("key", key).Int("count", len(list)).
			Msg("Successfully loaded JSON list")
		return list
	case []string:
		log.Debug().Str("path", path).Str("key", key).Int("count", len(v)).
			Msg("Successfully loaded JSON list")
		return v
	default:
		log.Warn().Str("path", path).Str("key", key).
			Msg("Value is not an array, returning empty list")
		return []string{}
	}
}
