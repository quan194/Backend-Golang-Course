package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type HttpConfig struct {
	Env EnvType `env:"ENV"`

	Host string `env:"HTTP_HOST"`
	Port int    `env:"HTTP_PORT"`

	GrpcHost string `env:"GRPC_HOST"`
	GrpcPort int    `env:"GRPC_PORT"`

	JwtKey string `env:"JWT_KEY"`
}

// LoadHttpConfig loads config based on the environment.
// envType: "local" (default) or "prod"
func LoadHttpConfig() (*HttpConfig, error) {
	envType := getEnvType()

	if envType == EnvTypeLocal { // if local, inject env vars from local .env file
		if err := godotenv.Load(".env"); err != nil {
			return nil, fmt.Errorf("failed to load local env file: %s", err)
		}
	}

	// parse config from env vars
	cfg := new(HttpConfig)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return cfg, nil
}
