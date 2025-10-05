package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type GrpcConfig struct {
	Env EnvType `env:"ENV"`

	Host string `env:"GRPC_HOST"`
	Port int    `env:"GRPC_PORT"`

	DatabaseUser     string `env:"DATABASE_USER"`
	DatabasePassword string `env:"DATABASE_PASSWORD"`
	DatabaseHost     string `env:"DATABASE_HOST"`
	DatabasePort     int    `env:"DATABASE_PORT"`
	DatabaseName     string `env:"DATABASE_NAME"`

	RedisHost    string `env:"REDIS_HOST"`
	RedisPort    int    `env:"REDIS_PORT"`
	RedisEnabled bool   `env:"REDIS_ENABLED"`

	KafkaBrokers []string `env:"KAFKA_BROKERS"`
	KafkaTopic   string   `env:"KAFKA_TOPIC"`
}

func LoadGrpcConfig() (*GrpcConfig, error) {
	envType := getEnvType()

	if envType == EnvTypeLocal { // if local, inject env vars from local .env file
		if err := godotenv.Load(".env"); err != nil {
			return nil, fmt.Errorf("failed to load local env file: %s", err)
		}
	}

	// parse config from env vars
	cfg := new(GrpcConfig)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return cfg, nil
}
