package config

import (
	"os"

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
	SentryDSN string `env:"SENTRY_DSN"`

	UploadDir string `env:"UPLOAD_DIR" envDefault:"./uploads"`

	// Midtrans payment gateway
	MidtransServerKey string `env:"MIDTRANS_SERVER_KEY"`
	MidtransSandbox   bool   `env:"MIDTRANS_SANDBOX" envDefault:"true"`

	// Cloudflare R2 (leave empty to use local filesystem)
	R2AccountID       string `env:"R2_ACCOUNT_ID"`
	R2AccessKeyID     string `env:"R2_ACCESS_KEY_ID"`
	R2SecretAccessKey string `env:"R2_SECRET_ACCESS_KEY"`
	R2BucketName      string `env:"R2_BUCKET_NAME"`
	R2PublicURL       string `env:"R2_PUBLIC_URL"`
}

var cfg Config

func InitConfig() {
	envFile := ".env.local"
	if os.Getenv("ENV") == "production" {
		envFile = ".env.production"
	}
	err := godotenv.Load(envFile)
	if err != nil {
		logger.WithError(err).Warn("unable to load env file: ", envFile)
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
