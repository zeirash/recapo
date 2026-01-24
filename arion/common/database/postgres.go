package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/logger"
)

var db *sql.DB

type DbConfig struct {
	User     string `required:"true" split_words:"true"`
	Password string `required:"true" split_words:"true"`
	Host     string `required:"true" split_words:"true"`
	Port     int    `required:"true" split_words:"true"`
	Name     string `required:"true" split_words:"true"`

	disableSSL bool
}

func url(dbConfig DbConfig) string {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)

	if dbConfig.disableSSL {
		url = fmt.Sprintf("%s?sslmode=disable", url)
	}

	return url
}

func connect(conn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// InitDB initializes the database connection pool. Should be called once at startup.
func InitDB() {
	cfg := config.GetConfig()
	conn := url(DbConfig{
		User:     cfg.DbUsername,
		Password: cfg.DbPassword,
		Host:     cfg.DbHost,
		Port:     cfg.DbPort,
		Name:     cfg.DbName,

		disableSSL: true,
	})

	var err error
	db, err = connect(conn)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to database")
	}

	logger.Info("Database connection established")
}

// GetDB returns the database connection pool.
func GetDB() *sql.DB {
	if db == nil {
		logger.Fatal("database not initialized - call InitDB() first")
	}
	return db
}

// CloseDB closes the database connection. Should be called on shutdown.
func CloseDB() {
	if db != nil {
		db.Close()
	}
}
