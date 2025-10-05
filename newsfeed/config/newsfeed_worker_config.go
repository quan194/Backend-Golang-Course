package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type NewsfeedWorkerConfig struct {
	Env EnvType `env:"ENV"`
	
	RedisHost string `env:"REDIS_HOST"`
	RedisPort int    `env:"REDIS_PORT"`

	KafkaBrokers       []string `env:"KAFKA_BROKERS"`
	KafkaTopic         string   `env:"KAFKA_TOPIC"`
	KafkaConsumerGroup string   `env:"KAFKA_CONSUMER_GROUP"`
}

func LoadNewsfeedWorkerConfig() (*NewsfeedWorkerConfig, error) {
	envType := getEnvType()

	if envType == EnvTypeLocal { // if local, inject env vars from local .env file
		if err := godotenv.Load(".env"); err != nil {
			return nil, fmt.Errorf("failed to load local env file: %s", err)
		}
	}

	// parse config from env vars
	cfg := new(NewsfeedWorkerConfig)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return cfg, nil
}
