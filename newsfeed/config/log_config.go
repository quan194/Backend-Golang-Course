package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type LogConfig struct {
	Type     string `env:"LOG_TYPE"`
	Level    string `env:"LOG_LEVEL"`
	Output   string `env:"LOG_OUTPUT"`
	Filename string `env:"LOG_FILENAME"`
	UseJSON  bool   `env:"LOG_USE_JSON"`
}

func LoadLogConfig() (*LogConfig, error) {
	envType := getEnvType()

	if envType == EnvTypeLocal { // if local, inject env vars from local .env file
		if err := godotenv.Load(".env"); err != nil {
			return nil, fmt.Errorf("failed to load local env file: %s", err)
		}
	}

	// parse config from env vars
	cfg := new(LogConfig)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return cfg, nil
}
