package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/zeirash/recapo/arion/common/logger"
)

type Config struct {
	Env         string `env:"ENV"`
	ServiceName string `env:"SERVICE_NAME"`
	ServicePort int    `env:"SERVICE_PORT"`
	Version     string `env:"VERSION"`

	// DB related configuration
	DbName     string `env:"DB_NAME"`
	DbHost     string `env:"DB_HOST"`
	DbUsername string `env:"DB_USERNAME"`
	DbPassword string `env:"DB_PASSWORD"`
	DbPort     int    `env:"DB_PORT"`

	SecretKey string `env:"SECRET_KEY"`
}

var cfg Config

func InitConfig() {
	// Loading the environment variables from '.env' file.
	err := godotenv.Load()
	if err != nil {
		logger.WithError(err).Warn("unable to load .env file")
		return
	}

	err = env.Parse(&cfg) // Parse environment variables into `Config`
	if err != nil {
		logger.WithError(err).Error("unable to parse environment variables")
	}
}

func GetConfig() Config {
	return cfg
}
