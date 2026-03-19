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
	Port        int    `env:"PORT"`
	Version     string `env:"VERSION"`

	// DB related configuration
	DbName     string `env:"DB_NAME"`
	DbHost     string `env:"DB_HOST"`
	DbUsername string `env:"DB_USERNAME"`
	DbPassword string `env:"DB_PASSWORD"`
	DbPort     int    `env:"DB_PORT"`

	SecretKey   string `env:"SECRET_KEY"`
	SentryDSN   string `env:"SENTRY_DSN"`
	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"http://localhost:3000,http://localhost:3001"`

	UploadDir string `env:"UPLOAD_DIR" envDefault:"./uploads"`

	// Midtrans payment gateway
	MidtransServerKey string `env:"MIDTRANS_SERVER_KEY"`
	MidtransBaseURL   string `env:"MIDTRANS_BASE_URL" envDefault:"https://app.sandbox.midtrans.com"`

	FrontendURL string `env:"MIDTRANS_FRONTEND_URL" envDefault:"http://localhost:3000"`

	// SMTP (leave empty to use development log mode)
	SMTPHost string `env:"SMTP_HOST"`
	SMTPPort int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUser string `env:"SMTP_USER"`
	SMTPPass string `env:"SMTP_PASS"`

	// Resend (preferred over SMTP when set)
	ResendAPIKey    string `env:"RESEND_API_KEY"`
	ResendFromEmail string `env:"RESEND_FROM_EMAIL"`
	ResendFromName  string `env:"RESEND_FROM_NAME" envDefault:"Recapo"`

	// Cloudflare R2 (leave empty to use local filesystem)
	R2AccountID       string `env:"R2_ACCOUNT_ID"`
	R2AccessKeyID     string `env:"R2_ACCESS_KEY_ID"`
	R2SecretAccessKey string `env:"R2_SECRET_ACCESS_KEY"`
	R2BucketName      string `env:"R2_BUCKET_NAME"`
	R2PublicURL       string `env:"R2_PUBLIC_URL"`

	// GitHub (for feedback issues)
	GitHubToken     string `env:"GITHUB_TOKEN"`
	GitHubRepoOwner string `env:"GITHUB_REPO_OWNER"`
	GitHubRepoName  string `env:"GITHUB_REPO_NAME"`
}

var cfg Config

func InitConfig() {
	envFile := ".env.local"
	if os.Getenv("ENV") == "production" {
		envFile = ".env.production"
	}
	if err := godotenv.Load(envFile); err != nil {
		logger.WithError(err).Warn("unable to load env file: ", envFile)
	}

	err := env.Parse(&cfg) // Parse environment variables into `Config`
	if err != nil {
		logger.WithError(err).Error("unable to parse environment variables")
	}
}

func GetConfig() Config {
	return cfg
}
