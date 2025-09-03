package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// AppConfig holds all configuration variables
type AppConfig struct {
	// Database
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     string `envconfig:"DB_PORT" default:"3306"`
	DBUser     string `envconfig:"DB_USER" default:"root"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"password"`
	DBName     string `envconfig:"DB_NAME" default:"muzzapp"`

	// Redis
	RedisHost     string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`

	// gRPC
	GRPCPort string `envconfig:"GRPC_PORT" default:"50051"`

	// pagination size limit
	PaginationSize int64 `envconfig:"PAGINATION_SIZE" default:"50"`
}

// Load reads environment variables into AppConfig
func Load() *AppConfig {
	var cfg AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to load config from environment: %v", err)
	}
	return &cfg
}
