package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
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
		fmt.Printf("unable to load .env file: %e", err)
		return
	}

	err = env.Parse(&cfg) // Parse environment variables into `Config`
	if err != nil {
		fmt.Printf("unable to parse ennvironment variables: %e", err)
	}
}

func GetConfig() Config {
	return cfg
}
